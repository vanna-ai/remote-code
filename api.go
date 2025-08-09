package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
    "path/filepath"
	"strconv"
	"strings"
	"time"
	"remote-code/db"
	
	"github.com/robert-nix/ansihtml"
)

// -----------------
// Git API utilities
// -----------------

type GitFile struct {
    Path    string `json:"path"`
    X       string `json:"x"`     // Index status
    Y       string `json:"y"`     // Worktree status
    Staged  bool   `json:"staged"`
}

type GitStatus struct {
    CurrentBranch  string    `json:"currentBranch"`
    Upstream       string    `json:"upstream"`
    Ahead          int       `json:"ahead"`
    Behind         int       `json:"behind"`
    IsDirty        bool      `json:"isDirty"`
    StagedFiles    []GitFile `json:"stagedFiles"`
    UnstagedFiles  []GitFile `json:"unstagedFiles"`
    UntrackedFiles []GitFile `json:"untrackedFiles"`
    MergeConflicts []GitFile `json:"mergeConflicts"`
}

func runGit(dir string, args ...string) (string, int, error) {
    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    out, err := cmd.CombinedOutput()
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            return string(out), exitErr.ExitCode(), err
        }
        return string(out), -1, err
    }
    return string(out), 0, nil
}

func parsePorcelainStatus(output string) (staged, unstaged, untracked, conflicts []GitFile) {
    lines := strings.Split(output, "\x00")
    for _, line := range lines {
        if line == "" {
            continue
        }

        // Porcelain -z format: XY<space>PATH (optionally with rename -> PATH\tPATH)
        if len(line) < 3 {
            continue
        }
        x := string(line[0])
        y := string(line[1])
        rest := strings.TrimSpace(line[2:])
        path := rest
        if strings.Contains(rest, "\t") {
            parts := strings.SplitN(rest, "\t", 2)
            path = parts[1] // show new path for renames
        }

        gf := GitFile{Path: path, X: x, Y: y}

        // Untracked
        if x == "?" && y == "?" {
            untracked = append(untracked, gf)
            continue
        }

        // Conflicts: both modified or special states (e.g., UU, AA, DD)
        if (x == "U" || y == "U") || (x == "A" && y == "A") || (x == "D" && y == "D") {
            conflicts = append(conflicts, gf)
            continue
        }

        // Staged if index has changes
        if x != " " && x != "?" {
            gf.Staged = true
            staged = append(staged, gf)
        }

        // Unstaged if worktree has changes
        if y != " " && y != "?" {
            gf.Staged = false
            unstaged = append(unstaged, gf)
        }
    }
    return
}

func getGitStatus(dir string) (*GitStatus, int, string, error) {
    // Branch and ahead/behind via -sb
    short, _, err := runGit(dir, "status", "-sb")
    if err != nil {
        return nil, 1, short, err
    }

    // Files via porcelain -z
    porcelain, _, err := runGit(dir, "status", "--porcelain=1", "-z")
    if err != nil {
        return nil, 1, porcelain, err
    }

    status := &GitStatus{StagedFiles: []GitFile{}, UnstagedFiles: []GitFile{}, UntrackedFiles: []GitFile{}, MergeConflicts: []GitFile{}}

    // Parse branch line: e.g., "## main...origin/main [ahead 1]"
    for _, line := range strings.Split(strings.TrimSpace(short), "\n") {
        if strings.HasPrefix(line, "## ") {
            meta := strings.TrimPrefix(line, "## ")
            // Split branch...upstream and position info
            parts := strings.Split(meta, "...")
            status.CurrentBranch = strings.Split(parts[0], " ")[0]
            if len(parts) > 1 {
                rest := parts[1]
                segs := strings.SplitN(rest, " ", 2)
                status.Upstream = strings.TrimSpace(segs[0])
                if len(segs) > 1 {
                    pos := segs[1]
                    if strings.Contains(pos, "ahead ") {
                        fmt.Sscanf(pos, "[ahead %d]", &status.Ahead)
                    }
                    if strings.Contains(pos, "behind ") {
                        // supports formats like [ahead 1, behind 2]
                        fmt.Sscanf(pos, "[behind %d]", &status.Behind)
                        var a, b int
                        if strings.Contains(pos, "ahead ") && strings.Contains(pos, "behind ") {
                            fmt.Sscanf(pos, "[ahead %d, behind %d]", &a, &b)
                            status.Ahead = a
                            status.Behind = b
                        }
                    }
                }
            }
        }
    }

    staged, unstaged, untracked, conflicts := parsePorcelainStatus(porcelain)
    status.StagedFiles = staged
    status.UnstagedFiles = unstaged
    status.UntrackedFiles = untracked
    status.MergeConflicts = conflicts
    status.IsDirty = len(staged) > 0 || len(unstaged) > 0 || len(untracked) > 0 || len(conflicts) > 0

    return status, 0, "", nil
}

func handleGitAPI(w http.ResponseWriter, r *http.Request, _ context.Context, pathParts []string) {
    if len(pathParts) == 0 {
        json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid git endpoint"})
        return
    }

    // For GET endpoints we accept ?path=...; for POST we read from JSON body.
    // Each branch validates presence of path and returns JSON-form errors.

    switch pathParts[0] {
    case "status":
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        dir := r.URL.Query().Get("path")
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        st, code, stdout, err := getGitStatus(dir)
        if err != nil && st == nil {
            // Likely not a repo or other error
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "error":   err.Error(),
                "code":    code,
                "output":  stdout,
            })
            return
        }
        json.NewEncoder(w).Encode(st)

    case "diff":
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        dir := r.URL.Query().Get("path")
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        file := r.URL.Query().Get("file")
        staged := r.URL.Query().Get("staged") == "true"
        args := []string{"diff"}
        if staged {
            args = append(args, "--staged")
        }
        if file != "" {
            args = append(args, "--", file)
        }
        out, _, err := runGit(dir, args...)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "output": out})
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"diff": out})

    case "add":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path string `json:"path"`
            File string `json:"file"`
            All  bool   `json:"all"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON"})
            return
        }
        dir := body.Path
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        var args []string
        if body.All {
            args = []string{"add", "-A"}
        } else if body.File != "" {
            args = []string{"add", "--", body.File}
        } else {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing file or all flag"})
            return
        }
        out, _, err := runGit(dir, args...)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "output": out})
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

    case "unstage":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path string `json:"path"`
            File string `json:"file"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.File == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: requires file"})
            return
        }
        dir := body.Path
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        out, _, err := runGit(dir, "restore", "--staged", "--", body.File)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "output": out})
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

    case "commit":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path    string `json:"path"`
            Message string `json:"message"`
            Amend   bool   `json:"amend"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Message == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: requires message"})
            return
        }
        dir := body.Path
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        args := []string{"commit", "-m", body.Message}
        if body.Amend {
            args = append(args, "--amend")
        }
        out, code, err := runGit(dir, args...)
        if err != nil {
            // No changes to commit or other error
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "code": code, "output": out})
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "output": out})

    case "merge":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path   string `json:"path"`
            Branch string `json:"branch"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: requires branch"})
            return
        }
        worktreeDir := body.Path
        if worktreeDir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }

        // Resolve base repository path from any worktree dir via git-common-dir
        commonDirOut, _, err := runGit(worktreeDir, "rev-parse", "--git-common-dir")
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a git repo"})
            return
        }
        gitDir := strings.TrimSpace(commonDirOut)
        basePath := filepath.Dir(gitDir)

        // Step 1: Ensure we are on main
        headOut, _, err := runGit(basePath, "symbolic-ref", "--quiet", "--short", "HEAD")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "check_branch",
                "error":  "Failed to detect current branch",
                "output": headOut,
            })
            return
        }
        currentBranch := strings.TrimSpace(headOut)
        if currentBranch != "main" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":    false,
                "step":  "check_branch",
                "error": fmt.Sprintf("Base repo not on main (on %s)", currentBranch),
            })
            return
        }

        // Step 2: Ensure working tree is clean
        statusOut, _, err := runGit(basePath, "status", "--porcelain")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "check_clean",
                "error":  "Failed to check status",
                "output": statusOut,
            })
            return
        }
        if strings.TrimSpace(statusOut) != "" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "check_clean",
                "error":  "Working tree not clean on main",
                "output": statusOut,
            })
            return
        }

        // Step 3: Fetch origin
        fetchOut, code, err := runGit(basePath, "fetch", "origin")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "fetch",
                "code":   code,
                "error":  "git fetch failed",
                "output": fetchOut,
            })
            return
        }

        // Step 4: Merge fast-forward only into main
        mergeOut, code, err := runGit(basePath, "merge", "--ff-only", body.Branch)
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "merge",
                "code":   code,
                "error":  "git merge failed",
                "output": mergeOut,
            })
            return
        }

        json.NewEncoder(w).Encode(map[string]interface{}{
            "ok":     true,
            "output": mergeOut,
        })

    case "push":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path string `json:"path"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON"})
            return
        }
        worktreeDir := body.Path
        if worktreeDir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }

        // Resolve base repository path
        commonDirOut, _, err := runGit(worktreeDir, "rev-parse", "--git-common-dir")
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a git repo"})
            return
        }
        gitDir := strings.TrimSpace(commonDirOut)
        basePath := filepath.Dir(gitDir)

        // Ensure we are on main
        headOut, _, err := runGit(basePath, "symbolic-ref", "--quiet", "--short", "HEAD")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "check_branch",
                "error":  "Failed to detect current branch",
                "output": headOut,
            })
            return
        }
        currentBranch := strings.TrimSpace(headOut)
        if currentBranch != "main" {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":    false,
                "step":  "check_branch",
                "error": fmt.Sprintf("Base repo not on main (on %s)", currentBranch),
            })
            return
        }

        // Ensure there is an upstream
        upstreamOut, _, err := runGit(basePath, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "check_upstream",
                "error":  "No upstream configured for main",
            })
            return
        }
        _ = strings.TrimSpace(upstreamOut) // e.g., origin/main

        // Push to upstream
        pushOut, code, err := runGit(basePath, "push")
        if err != nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "ok":     false,
                "step":   "push",
                "code":   code,
                "error":  "git push failed",
                "output": pushOut,
            })
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "output": pushOut})

    case "branches":
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        dir := r.URL.Query().Get("path")
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        includeRemotes := r.URL.Query().Get("includeRemotes") == "true"
        out, _, err := runGit(dir, "branch", "--format", "%(refname:short)")
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to list branches"})
            return
        }
        var branches []string
        for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
            if strings.TrimSpace(line) != "" {
                branches = append(branches, strings.TrimSpace(line))
            }
        }
        if includeRemotes {
            outR, _, err := runGit(dir, "branch", "-r", "--format", "%(refname:short)")
            if err == nil {
                for _, line := range strings.Split(strings.TrimSpace(outR), "\n") {
                    n := strings.TrimSpace(line)
                    if n == "" {
                        continue
                    }
                    if n == "origin/HEAD" { // skip symbolic default pointer
                        continue
                    }
                    branches = append(branches, n)
                }
            }
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"branches": branches})

    case "checkout":
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        var body struct{
            Path   string `json:"path"`
            Branch string `json:"branch"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: requires branch"})
            return
        }
        dir := body.Path
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        out, _, err := runGit(dir, "checkout", body.Branch)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error(), "output": out})
            return
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})

    case "base-branch":
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
            return
        }
        dir := r.URL.Query().Get("path")
        if dir == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
            return
        }
        // Resolve base repository path from any worktree dir via git-common-dir
        commonDirOut, _, err := runGit(dir, "rev-parse", "--git-common-dir")
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a git repo"})
            return
        }
        gitDir := strings.TrimSpace(commonDirOut)
        basePath := filepath.Dir(gitDir)
        // List worktrees in base repo and find the base worktree entry
        wtOut, _, err := runGit(basePath, "worktree", "list", "--porcelain")
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to list worktrees"})
            return
        }
        var currentWT string
        var baseBranch string
        for _, line := range strings.Split(wtOut, "\n") {
            if strings.HasPrefix(line, "worktree ") {
                currentWT = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
            } else if strings.HasPrefix(line, "branch ") && currentWT != "" {
                br := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
                br = strings.TrimPrefix(br, "refs/heads/")
                if currentWT == basePath {
                    baseBranch = br
                    break
                }
            }
        }
        if baseBranch == "" {
            // Fallback: try symbolic-ref on base path
            ref, _, err := runGit(basePath, "symbolic-ref", "--quiet", "--short", "HEAD")
            if err == nil {
                baseBranch = strings.TrimSpace(ref)
            }
        }
        json.NewEncoder(w).Encode(map[string]interface{}{"basePath": basePath, "branch": baseBranch})

    default:
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unknown git action"})
    }
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the path
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	pathParts := strings.Split(path, "/")

	if len(pathParts) == 0 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	switch pathParts[0] {
	case "dashboard":
		handleDashboardAPI(w, r, ctx, pathParts[1:])
	case "roots":
		handleRootsAPI(w, r, ctx, pathParts[1:])
	case "projects":
		handleProjectsAPI(w, r, ctx, pathParts[1:])
	case "agents":
		handleAgentsAPI(w, r, ctx, pathParts[1:])
	case "base-directories":
		handleBaseDirectoriesAPI(w, r, ctx, pathParts[1:])
	case "worktrees":
		handleWorktreesAPI(w, r, ctx, pathParts[1:])
	case "tasks":
		handleTasksAPI(w, r, ctx, pathParts[1:])
	case "task-executions":
		handleTaskExecutionsAPI(w, r, ctx, pathParts[1:])
	case "tmux-sessions":
		handleTmuxSessionsAPI(w, r, ctx, pathParts[1:])
	case "git":
		handleGitAPI(w, r, ctx, pathParts[1:])
	default:
		http.Error(w, "Unknown API endpoint", http.StatusNotFound)
	}
}

type DashboardStats struct {
	ActiveSessions   int `json:"active_sessions"`
	Projects         int `json:"projects"`
	TaskExecutions   int `json:"task_executions"`
	Agents          int `json:"agents"`
}

func handleDashboardAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) > 0 && pathParts[0] == "stats" {
			// Get dashboard statistics
			stats := DashboardStats{}
			
			// Count active tmux sessions
			cmd := exec.Command("tmux", "list-sessions")
			output, err := cmd.Output()
			if err == nil {
				sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
				if len(sessions) > 0 && sessions[0] != "" {
					stats.ActiveSessions = len(sessions)
				}
			}
			
			// Count projects
			projects, err := queries.ListProjects(ctx)
			if err == nil {
				stats.Projects = len(projects)
			}
			
			// Count task executions
			executions, err := queries.ListTaskExecutions(ctx)
			if err == nil {
				stats.TaskExecutions = len(executions)
			}
			
			// Count agents
			agents, err := queries.ListAgents(ctx)
			if err == nil {
				stats.Agents = len(agents)
			}
			
			json.NewEncoder(w).Encode(stats)
		} else {
			http.Error(w, "Unknown dashboard endpoint", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRootsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all roots - for now just return empty array
			json.NewEncoder(w).Encode([]Root{})
		} else {
			// Get specific root
			id, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid root ID", http.StatusBadRequest)
				return
			}

			root, err := queries.GetRoot(ctx, id)
			if err != nil {
				http.Error(w, "Root not found", http.StatusNotFound)
				return
			}

			// Get associated agents
			agents, err := queries.GetAgentsByRootID(ctx, id)
			if err != nil {
				http.Error(w, "Failed to get agents", http.StatusInternalServerError)
				return
			}

			// Get associated projects (simplified for now)
			projects := []Project{}

			result := dbRootToRoot(root, agents, projects)
			json.NewEncoder(w).Encode(result)
		}

	case "POST":
		var createReq struct {
			LocalPort   string  `json:"local_port"`
			ExternalUrl *string `json:"external_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		root, err := queries.CreateRoot(ctx, db.CreateRootParams{
			LocalPort:   createReq.LocalPort,
			ExternalUrl: stringToNullString(createReq.ExternalUrl),
		})
		if err != nil {
			http.Error(w, "Failed to create root", http.StatusInternalServerError)
			return
		}

		result := dbRootToRoot(root, []db.Agent{}, []Project{})
		json.NewEncoder(w).Encode(result)

	default:
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTmuxSessionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		// Get all tmux sessions
		sessions, err := getTmuxSessions()
		if err != nil {
			log.Printf("Failed to get tmux sessions: %v", err)
			http.Error(w, "Failed to get tmux sessions", http.StatusInternalServerError)
			return
		}
		
		json.NewEncoder(w).Encode(sessions)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type TmuxSession struct {
	Name      string `json:"name"`
	Created   string `json:"created"`
	Preview   string `json:"preview"`
	TaskID    *int64 `json:"task_id,omitempty"`
	TaskName  *string `json:"task_name,omitempty"`
	AgentID   *int64 `json:"agent_id,omitempty"`
	AgentName *string `json:"agent_name,omitempty"`
	IsTask    bool   `json:"is_task"`
}

func getTmuxSessions() ([]TmuxSession, error) {
	// Get list of tmux sessions
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}|#{session_created}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var sessions []TmuxSession
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}
		
		sessionName := parts[0]
		created := parts[1]
		
		// Get preview of the session with colors - capture more lines for scrollable view
		preview := ""
		previewCmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-e", "-p", "-S", "-20")
		if previewOutput, err := previewCmd.Output(); err == nil {
			rawPreview := strings.TrimSpace(string(previewOutput))
			// Convert ANSI to HTML
			preview = string(ansihtml.ConvertToHTML([]byte(rawPreview)))
		}
		
		session := TmuxSession{
			Name:    sessionName,
			Created: created,
			Preview: preview,
		}
		
		// Check if this is a task session
		if strings.HasPrefix(sessionName, "task_") {
			session.IsTask = true
			// Parse task and agent IDs from session name (format: task_{taskId}_agent_{agentId})
			parts := strings.Split(sessionName, "_")
			if len(parts) >= 4 {
				if taskID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					session.TaskID = &taskID
					// Look up task name
					if task, err := queries.GetTask(context.Background(), taskID); err == nil {
						session.TaskName = &task.Title
					}
				}
				if agentID, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
					session.AgentID = &agentID
					// Look up agent name
					if agent, err := queries.GetAgent(context.Background(), agentID); err == nil {
						session.AgentName = &agent.Name
					}
				}
			}
		}
		
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}

func handleProjectsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	// Handle tasks sub-resource: /api/projects/{id}/tasks
	if len(pathParts) >= 2 && pathParts[1] == "tasks" {
		handleProjectTasksAPI(w, r, ctx, pathParts)
		return
	}
	
	// Handle base-directories sub-resource: /api/projects/{id}/base-directories
	if len(pathParts) >= 2 && pathParts[1] == "base-directories" {
		handleProjectBaseDirectoriesAPI(w, r, ctx, pathParts)
		return
	}
	
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all projects
			dbProjects, err := queries.GetProjectsByRootID(ctx, 1) // Use default root_id for now
			if err != nil {
				// If root doesn't exist, return empty array instead of error
				log.Printf("No projects found for root_id 1: %v", err)
				json.NewEncoder(w).Encode([]Project{})
				return
			}

			projects := make([]Project, 0)
			for _, dbProject := range dbProjects {
				// Get base directories for this project
				dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, dbProject.ID)
				if err != nil {
					// Log error but continue
					log.Printf("Failed to get base directories for project %d: %v", dbProject.ID, err)
					dbBaseDirs = []db.BaseDirectory{}
				}

				var baseDirs []BaseDirectory
				for _, dbBaseDir := range dbBaseDirs {
					baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
				}

				// Get tasks for this project
				tasks, err := loadTasksForProject(ctx, dbProject.ID, baseDirs)
				if err != nil {
					log.Printf("Failed to get tasks for project %d: %v", dbProject.ID, err)
					tasks = []Task{}
				}

				// Ensure arrays are never null
				if baseDirs == nil {
					baseDirs = []BaseDirectory{}
				}
				if tasks == nil {
					tasks = []Task{}
				}
				
				project := Project{
					ID:              dbProject.ID,
					Name:            dbProject.Name,
					BaseDirectories: baseDirs,
					Tasks:           tasks,
				}
				projects = append(projects, project)
			}

			json.NewEncoder(w).Encode(projects)
		} else {
			// Get specific project
			id, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid project ID", http.StatusBadRequest)
				return
			}

			project, err := queries.GetProject(ctx, id)
			if err != nil {
				http.Error(w, "Project not found", http.StatusNotFound)
				return
			}

			// Get base directories
			dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, id)
			if err != nil {
				http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
				return
			}

			var baseDirs []BaseDirectory
			for _, dbBaseDir := range dbBaseDirs {
				baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
			}

			// Get tasks for this project
			tasks, err := loadTasksForProject(ctx, id, baseDirs)
			if err != nil {
				log.Printf("Failed to get tasks for project %d: %v", id, err)
				tasks = []Task{}
			}

			// Ensure arrays are never null
			if baseDirs == nil {
				baseDirs = []BaseDirectory{}
			}
			if tasks == nil {
				tasks = []Task{}
			}

			result := Project{
				ID:              project.ID,
				Name:            project.Name,
				BaseDirectories: baseDirs,
				Tasks:           tasks,
			}
			json.NewEncoder(w).Encode(result)
		}

	case "POST":
		var createReq struct {
			RootId int64  `json:"root_id"`
			Name   string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		project, err := queries.CreateProject(ctx, db.CreateProjectParams{
			RootID: createReq.RootId,
			Name:   createReq.Name,
		})
		if err != nil {
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		result := Project{
			ID:              project.ID,
			Name:            project.Name,
			BaseDirectories: []BaseDirectory{},
			Tasks:           []Task{},
		}
		json.NewEncoder(w).Encode(result)

	case "DELETE":
		if len(pathParts) == 0 {
			http.Error(w, "Project ID required", http.StatusBadRequest)
			return
		}
		
		projectID, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}
		
		// Delete all tasks in this project first (which will also clean up task executions)
		tasks, err := queries.GetTasksByProjectID(ctx, projectID)
		if err == nil {
			for _, task := range tasks {
				// Check if there are any active task executions for this task
				executions, err := queries.GetTaskExecutionsByTaskID(ctx, task.ID)
				if err == nil && len(executions) > 0 {
					// Delete all task executions first
					for _, execution := range executions {
						err := deleteTaskExecutionWithCleanup(ctx, execution.ID)
						if err != nil {
							log.Printf("Warning: failed to cleanup task execution %d: %v", execution.ID, err)
						}
					}
				}
				
				// Delete the task
				err = queries.DeleteTask(ctx, task.ID)
				if err != nil {
					log.Printf("Warning: failed to delete task %d: %v", task.ID, err)
				}
			}
		}
		
		// Delete all base directories for this project
		baseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err == nil {
			for _, baseDir := range baseDirs {
				err = queries.DeleteBaseDirectory(ctx, baseDir.ID)
				if err != nil {
					log.Printf("Warning: failed to delete base directory %d: %v", baseDir.ID, err)
				}
			}
		}
		
		// Finally delete the project
		err = queries.DeleteProject(ctx, projectID)
		if err != nil {
			log.Printf("Failed to delete project: %v", err)
			http.Error(w, "Failed to delete project", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Placeholder handlers for other endpoints
func handleAgentsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	// Handle specific agent operations: /api/agents/{id}
	if len(pathParts) > 0 && pathParts[0] != "detect" {
		handleSingleAgentAPI(w, r, ctx, pathParts)
		return
	}
	
	switch r.Method {
	case "GET":
		if len(pathParts) > 0 && pathParts[0] == "detect" {
			// Detect available agents on the system
			handleAgentDetection(w, r, ctx)
			return
		}
		
		// List all agents for default root (assuming root_id = 1 for now)
		dbAgents, err := queries.GetAgentsByRootID(ctx, 1)
		if err != nil {
			log.Printf("Failed to get agents: %v", err)
			json.NewEncoder(w).Encode([]Agent{})
			return
		}
		
		agents := make([]Agent, 0)
		for _, dbAgent := range dbAgents {
			agents = append(agents, Agent{
				ID:      dbAgent.ID,
				Name:    dbAgent.Name,
				Command: dbAgent.Command,
				Params:  dbAgent.Params,
			})
		}
		
		json.NewEncoder(w).Encode(agents)
		
	case "POST":
		// Create a new agent
		var createReq struct {
			RootId  int64  `json:"root_id"`
			Name    string `json:"name"`
			Command string `json:"command"`
			Params  string `json:"params"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		agent, err := queries.CreateAgent(ctx, db.CreateAgentParams{
			RootID:  createReq.RootId,
			Name:    createReq.Name,
			Command: createReq.Command,
			Params:  createReq.Params,
		})
		if err != nil {
			log.Printf("Failed to create agent: %v", err)
			http.Error(w, "Failed to create agent", http.StatusInternalServerError)
			return
		}
		
		result := Agent{
			ID:      agent.ID,
			Name:    agent.Name,
			Command: agent.Command,
			Params:  agent.Params,
		}
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleAgentDetection(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Define the agents to detect
	agentsToDetect := []string{"claude", "amp", "codex", "gemini", "codebuff", "aider", "opencode", "friday", "grok"}
	
	var availableAgents []map[string]interface{}
	
	for _, agentName := range agentsToDetect {
		cmd := exec.Command("which", agentName)
		output, err := cmd.Output()
		
		if err == nil && len(output) > 0 {
			// Agent is available
			path := strings.TrimSpace(string(output))
			availableAgents = append(availableAgents, map[string]interface{}{
				"name":      agentName,
				"command":   agentName,
				"path":      path,
				"available": true,
			})
		} else {
			// Agent not found
			availableAgents = append(availableAgents, map[string]interface{}{
				"name":      agentName,
				"command":   agentName,
				"path":      "",
				"available": false,
			})
		}
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": availableAgents,
	})
}

func handleSingleAgentAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) == 0 {
		http.Error(w, "Agent ID required", http.StatusBadRequest)
		return
	}
	
	agentID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "PUT":
		// Update agent
		var updateReq struct {
			Name    string `json:"name"`
			Command string `json:"command"`
			Params  string `json:"params"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		updatedAgent, err := queries.UpdateAgent(ctx, db.UpdateAgentParams{
			ID:      agentID,
			Name:    updateReq.Name,
			Command: updateReq.Command,
			Params:  updateReq.Params,
		})
		if err != nil {
			log.Printf("Failed to update agent: %v", err)
			http.Error(w, "Failed to update agent", http.StatusInternalServerError)
			return
		}
		
		result := Agent{
			ID:      updatedAgent.ID,
			Name:    updatedAgent.Name,
			Command: updatedAgent.Command,
			Params:  updatedAgent.Params,
		}
		json.NewEncoder(w).Encode(result)
		
	case "DELETE":
		// Delete agent
		err := queries.DeleteAgent(ctx, agentID)
		if err != nil {
			log.Printf("Failed to delete agent: %v", err)
			http.Error(w, "Failed to delete agent", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTaskExecutionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	// Handle sub-endpoints like /api/task-executions/{id}/send-input
	if len(pathParts) >= 2 && pathParts[1] == "send-input" {
		handleSendInputToSession(w, r, ctx, pathParts)
		return
	}
	
	// Handle sub-endpoints like /api/task-executions/{id}/resend-task
	if len(pathParts) >= 2 && pathParts[1] == "resend-task" {
		handleResendTaskToSession(w, r, ctx, pathParts)
		return
	}
	
	switch r.Method {
	case "DELETE":
		// Handle task execution deletion
		if len(pathParts) > 0 {
			executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid execution ID", http.StatusBadRequest)
				return
			}
			
			err = deleteTaskExecutionWithCleanup(ctx, executionID)
			if err != nil {
				log.Printf("Failed to delete task execution: %v", err)
				http.Error(w, "Failed to delete task execution", http.StatusInternalServerError)
				return
			}
			
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		http.Error(w, "Task execution ID required", http.StatusBadRequest)
		
	case "GET":
		// Handle individual task execution by ID
		if len(pathParts) > 0 {
			executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid execution ID", http.StatusBadRequest)
				return
			}
			
			queries := db.New(database)
			execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
			if err != nil {
				log.Printf("Failed to get task execution: %v", err)
				http.Error(w, "Task execution not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(execution)
			return
		}
		
		// Get task executions, optionally filtered by task_id
		taskIDStr := r.URL.Query().Get("task_id")
		if taskIDStr != "" {
			taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			
			queries := db.New(database)
			executions, err := queries.GetTaskExecutionsByTaskID(ctx, taskID)
			if err != nil {
			log.Printf("Failed to get task executions: %v", err)
			http.Error(w, "Failed to get task executions", http.StatusInternalServerError)
			return
			}
			
			// Ensure we return empty array instead of null
		if executions == nil {
			executions = []db.GetTaskExecutionsByTaskIDRow{}
		}
		
		json.NewEncoder(w).Encode(executions)
			return
		}
		
		// Get all task executions
		queries := db.New(database)
		executions, err := queries.ListTaskExecutions(ctx)
		if err != nil {
			log.Printf("Failed to list task executions: %v", err)
			http.Error(w, "Failed to list task executions", http.StatusInternalServerError)
			return
		}
		
		// Ensure we return empty array instead of null
		if executions == nil {
			executions = []db.ListTaskExecutionsRow{}
		}
		
		json.NewEncoder(w).Encode(executions)
		
	case "POST":
		// Create and start a new task execution
		var createReq struct {
			TaskId  int64 `json:"task_id"`
			AgentId int64 `json:"agent_id"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Get the task details
		dbTask, err := queries.GetTask(ctx, createReq.TaskId)
		if err != nil {
			log.Printf("Failed to get task: %v", err)
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		
		// Get the agent details
		dbAgent, err := queries.GetAgent(ctx, createReq.AgentId)
		if err != nil {
			log.Printf("Failed to get agent: %v", err)
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}
		
		// Create a unique worktree path for this execution
		worktreePath := fmt.Sprintf("/tmp/task_%d_agent_%d_%d", createReq.TaskId, createReq.AgentId, time.Now().Unix())
		
		// Create the worktree
		dbWorktree, err := queries.CreateWorktree(ctx, db.CreateWorktreeParams{
			BaseDirectoryID: dbTask.BaseDirectoryID,
			Path:            worktreePath,
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
			ExternalUrl:     sql.NullString{Valid: false},
		})
		if err != nil {
			log.Printf("Failed to create worktree: %v", err)
			http.Error(w, "Failed to create worktree", http.StatusInternalServerError)
			return
		}
		
		// Create the task execution record
		dbTaskExecution, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:     createReq.TaskId,
			AgentID:    createReq.AgentId,
			WorktreeID: dbWorktree.ID,
			Status:     "starting",
		})
		if err != nil {
			log.Printf("Failed to create task execution: %v", err)
			http.Error(w, "Failed to create task execution", http.StatusInternalServerError)
			return
		}
		
		// Update task status to "in_progress" if it's not already
		if dbTask.Status != "in_progress" {
			_, err = queries.UpdateTask(ctx, db.UpdateTaskParams{
				ID:          dbTask.ID,
				Title:       dbTask.Title,
				Description: dbTask.Description,
				Status:      "in_progress",
			})
			if err != nil {
				log.Printf("Failed to update task status: %v", err)
			}
		}
		
		// Start the execution in the background
		go startTaskExecutionProcess(dbTaskExecution.ID, dbTask, dbAgent, dbWorktree)
		
		// Return the task execution details
		result := map[string]interface{}{
			"id":          dbTaskExecution.ID,
			"task_id":     dbTaskExecution.TaskID,
			"agent_id":    dbTaskExecution.AgentID,
			"worktree_id": dbTaskExecution.WorktreeID,
			"status":      dbTaskExecution.Status,
		}
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func startTaskExecutionProcess(executionID int64, task db.Task, agent db.Agent, worktree db.Worktree) {
	ctx := context.Background()
	
	log.Printf("Starting task execution %d: Task '%s' with agent '%s'", executionID, task.Title, agent.Name)
	
	// Get the base directory for setup commands
	baseDir, err := queries.GetBaseDirectoryByProjectAndID(ctx, db.GetBaseDirectoryByProjectAndIDParams{
		ProjectID:       task.ProjectID,
		BaseDirectoryID: worktree.BaseDirectoryID,
	})
	if err != nil {
		log.Printf("Failed to get base directory: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Create the worktree directory
	err = os.MkdirAll(worktree.Path, 0755)
	if err != nil {
		log.Printf("Failed to create worktree directory %s: %v", worktree.Path, err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Generate a unique tmux session name
	sessionName := fmt.Sprintf("task_%d_agent_%d", task.ID, agent.ID)
	
	// Start tmux session in the worktree directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", worktree.Path)
	err = tmuxCmd.Run()
	if err != nil {
		log.Printf("Failed to start tmux session: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Update worktree with tmux session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     sql.NullString{String: sessionName, Valid: true},
		DevServerTmuxID: worktree.DevServerTmuxID,
		ExternalUrl:     worktree.ExternalUrl,
	})
	if err != nil {
		log.Printf("Failed to update worktree with tmux session: %v", err)
	}
	
	// Function to send command and wait
	sendCommandAndWait := func(command string, description string) error {
		log.Printf("Executing %s: %s", description, command)
		cmd := exec.Command("tmux", "send-keys", "-t", sessionName, command, "Enter")
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to send %s command: %v", description, err)
		}
		
		// Wait a moment for command to execute
		time.Sleep(2 * time.Second)
		return nil
	}
	
	// Run git worktree setup if this is a git-initialized base directory
	if baseDir.GitInitialized {
		// Set up git worktree
		worktreeCommand := fmt.Sprintf("git worktree add %s", worktree.Path)
		// Change to base directory first, then set up worktree
		if err := sendCommandAndWait(fmt.Sprintf("cd %s", baseDir.Path), "change to base directory"); err != nil {
			log.Printf("Warning: %v", err)
		}
		if err := sendCommandAndWait(worktreeCommand, "git worktree setup"); err != nil {
			log.Printf("Warning: %v", err)
		}
		// Change back to worktree directory
		if err := sendCommandAndWait(fmt.Sprintf("cd %s", worktree.Path), "change to worktree directory"); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	
	// Run custom worktree setup commands if provided
	if baseDir.WorktreeSetupCommands != "" {
		if err := sendCommandAndWait(baseDir.WorktreeSetupCommands, "worktree setup commands"); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	
	// Start the agent command
	agentCommand := fmt.Sprintf("%s %s", agent.Command, agent.Params)
	if err := sendCommandAndWait(agentCommand, "agent command"); err != nil {
		log.Printf("Failed to start agent: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}
	
	// Update status to running
	updateTaskExecutionStatus(ctx, executionID, "running")
	
	// Wait 3 seconds for the agent to start, then send the task title and description
	go func() {
		time.Sleep(3 * time.Second)
		
		// Create the task prompt to send to the agent
		taskPrompt := fmt.Sprintf("Task: %s\n\nDescription: %s", task.Title, task.Description)
		
		// Send the task prompt to the agent session with agent-specific handling
		log.Printf("Sending initial task prompt to agent session: %s", taskPrompt)
		
		// Send the text first
		cmd := exec.Command("tmux", "send-keys", "-t", sessionName, taskPrompt)
		err := cmd.Run()
		if err != nil {
			log.Printf("Warning: Failed to send task prompt: %v", err)
			return
		}
		
		// Small delay for agent debouncing, then send Enter
		time.Sleep(100 * time.Millisecond)
		enterCmd := exec.Command("tmux", "send-keys", "-t", sessionName, "Enter")
		err = enterCmd.Run()
		if err != nil {
			log.Printf("Warning: Failed to send Enter for task prompt: %v", err)
		}
	}()
	
	log.Printf("Task execution %d started successfully in tmux session %s", executionID, sessionName)
}

func updateTaskExecutionStatus(ctx context.Context, executionID int64, status string) {
	_, err := queries.UpdateTaskExecutionStatus(ctx, db.UpdateTaskExecutionStatusParams{
		ID:     executionID,
		Status: status,
	})
	if err != nil {
		log.Printf("Failed to update task execution status: %v", err)
	}
}

func handleSendInputToSession(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if len(pathParts) < 1 {
		http.Error(w, "Task execution ID required", http.StatusBadRequest)
		return
	}
	
	executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}
	
	// Parse the request body
	var inputReq struct {
		Input string `json:"input"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&inputReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if inputReq.Input == "" {
		http.Error(w, "Input cannot be empty", http.StatusBadRequest)
		return
	}
	
	// Get the task execution to find the tmux session
	execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
	if err != nil {
		log.Printf("Failed to get task execution: %v", err)
		http.Error(w, "Task execution not found", http.StatusNotFound)
		return
	}
	
	// Get the worktree to find the tmux session
	worktree, err := queries.GetWorktree(ctx, execution.WorktreeID)
	if err != nil {
		log.Printf("Failed to get worktree: %v", err)
		http.Error(w, "Worktree not found", http.StatusNotFound)
		return
	}
	
	if !worktree.AgentTmuxID.Valid {
		http.Error(w, "No active tmux session for this task execution", http.StatusBadRequest)
		return
	}
	
	sessionName := worktree.AgentTmuxID.String
	
	// Send the input to the tmux session
	log.Printf("Sending input to session %s: %s", sessionName, inputReq.Input)
	
	// Send the text first
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, inputReq.Input)
	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to send input to tmux session: %v", err)
		http.Error(w, "Failed to send input to session", http.StatusInternalServerError)
		return
	}
	
	// Small delay for agent debouncing, then send Enter
	time.Sleep(100 * time.Millisecond)
	enterCmd := exec.Command("tmux", "send-keys", "-t", sessionName, "Enter")
	err = enterCmd.Run()
	if err != nil {
		log.Printf("Failed to send Enter to tmux session: %v", err)
		http.Error(w, "Failed to send Enter to session", http.StatusInternalServerError)
		return
	}
	
	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Input sent to session successfully",
		"session": sessionName,
	}
	json.NewEncoder(w).Encode(response)
}

func handleResendTaskToSession(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if len(pathParts) < 1 {
		http.Error(w, "Task execution ID required", http.StatusBadRequest)
		return
	}
	
	executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}
	
	// Get the task execution to find the task details and tmux session
	execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
	if err != nil {
		log.Printf("Failed to get task execution: %v", err)
		http.Error(w, "Task execution not found", http.StatusNotFound)
		return
	}
	
	// Get the task details
	task, err := queries.GetTask(ctx, execution.TaskID)
	if err != nil {
		log.Printf("Failed to get task: %v", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	
	// Get the worktree to find the tmux session
	worktree, err := queries.GetWorktree(ctx, execution.WorktreeID)
	if err != nil {
		log.Printf("Failed to get worktree: %v", err)
		http.Error(w, "Worktree not found", http.StatusNotFound)
		return
	}
	
	if !worktree.AgentTmuxID.Valid {
		http.Error(w, "No active tmux session for this task execution", http.StatusBadRequest)
		return
	}
	
	sessionName := worktree.AgentTmuxID.String
	
	// Create the task prompt to send to the agent
	taskPrompt := fmt.Sprintf("Task: %s\n\nDescription: %s", task.Title, task.Description)
	
	// Send the task prompt to the tmux session
	log.Printf("Re-sending task prompt to session %s", sessionName)
	
	// Send the text first
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, taskPrompt)
	err = cmd.Run()
	if err != nil {
		log.Printf("Failed to send task prompt to tmux session: %v", err)
		http.Error(w, "Failed to send task prompt to session", http.StatusInternalServerError)
		return
	}
	
	// Small delay for agent debouncing, then send Enter
	time.Sleep(100 * time.Millisecond)
	enterCmd := exec.Command("tmux", "send-keys", "-t", sessionName, "Enter")
	err = enterCmd.Run()
	if err != nil {
		log.Printf("Failed to send Enter to tmux session: %v", err)
		http.Error(w, "Failed to send Enter to session", http.StatusInternalServerError)
		return
	}
	
	// Return success response
	response := map[string]interface{}{
		"success":     true,
		"message":     "Task prompt re-sent to session successfully",
		"session":     sessionName,
		"task_title":  task.Title,
		"task_description": task.Description,
	}
	json.NewEncoder(w).Encode(response)
}

func handleBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Base directories API not implemented yet"})
}

func handleWorktreesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	queries := db.New(database)
	
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all worktrees - for now return empty array since no ListWorktrees method exists
			json.NewEncoder(w).Encode([]db.Worktree{})
		} else {
			// Get specific worktree
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			worktree, err := queries.GetWorktree(ctx, worktreeID)
			if err != nil {
				log.Printf("Failed to get worktree: %v", err)
				http.Error(w, "Worktree not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(worktree)
		}
		
	case "POST":
		// Handle sub-endpoints like /api/worktrees/{id}/dev-server
		if len(pathParts) >= 2 && pathParts[1] == "dev-server" {
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			// Start dev server for worktree
			err = startDevServer(ctx, queries, worktreeID)
			if err != nil {
				log.Printf("Failed to start dev server: %v", err)
				http.Error(w, "Failed to start dev server", http.StatusInternalServerError)
				return
			}
			
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server started"})
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
		
	case "DELETE":
		// Handle sub-endpoints like /api/worktrees/{id}/dev-server
		if len(pathParts) >= 2 && pathParts[1] == "dev-server" {
			worktreeID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid worktree ID", http.StatusBadRequest)
				return
			}
			
			// Stop dev server for worktree
			err = stopDevServer(ctx, queries, worktreeID)
			if err != nil {
				log.Printf("Failed to stop dev server: %v", err)
				http.Error(w, "Failed to stop dev server", http.StatusInternalServerError)
				return
			}
			
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server stopped"})
		} else {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		}
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func startDevServer(ctx context.Context, queries *db.Queries, worktreeID int64) error {
	worktree, err := queries.GetWorktree(ctx, worktreeID)
	if err != nil {
		return err
	}
	
	// We need to get the task to find dev server setup commands
	// First, let's get the task execution to find the task_id
	executions, err := queries.ListTaskExecutions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list task executions: %v", err)
	}
	
	var taskID int64
	for _, exec := range executions {
		if exec.WorktreeID == worktreeID {
			taskID = exec.TaskID
			break
		}
	}
	
	if taskID == 0 {
		return fmt.Errorf("no task found for worktree %d", worktreeID)
	}
	
	// Get task with base directory info to access dev server setup commands
	taskWithBaseDir, err := queries.GetTaskWithBaseDirectory(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task with base directory: %v", err)
	}
	
	// Create dev server session name
	devSessionName := fmt.Sprintf("dev_%d", worktreeID)
	
	// Start tmux session for dev server in the worktree directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", devSessionName, "-c", worktree.Path)
	err = tmuxCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to create dev server tmux session: %v", err)
	}
	
	// If there are dev server setup commands, execute them in the session
	if taskWithBaseDir.DevServerSetupCommands != "" {
		// Execute the dev server setup commands
		setupCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, taskWithBaseDir.DevServerSetupCommands, "Enter")
		err = setupCmd.Run()
		if err != nil {
			log.Printf("Failed to send dev server setup commands: %v", err)
			// Don't return error here, session is still created
		}
		
		// Add a separator and info message
		infoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "", "Enter")
		infoCmd.Run()
		
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server started. Session: "+devSessionName+"'", "Enter")
		echoCmd.Run()
	} else {
		// No setup commands, just show info
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server session created. No setup commands configured.'", "Enter")
		echoCmd.Run()
		
		bashCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "bash", "Enter")
		bashCmd.Run()
	}
	
	// Update worktree with dev server session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     worktree.AgentTmuxID,
		DevServerTmuxID: sql.NullString{String: devSessionName, Valid: true},
		ExternalUrl:     worktree.ExternalUrl,
	})
	
	return err
}

func stopDevServer(ctx context.Context, queries *db.Queries, worktreeID int64) error {
	worktree, err := queries.GetWorktree(ctx, worktreeID)
	if err != nil {
		return err
	}
	
	if worktree.DevServerTmuxID.Valid {
		// Kill the tmux session
		tmuxCmd := exec.Command("tmux", "kill-session", "-t", worktree.DevServerTmuxID.String)
		_ = tmuxCmd.Run() // Ignore error if session doesn't exist
	}
	
	// Update worktree to remove dev server session info
	_, err = queries.UpdateWorktree(ctx, db.UpdateWorktreeParams{
		ID:              worktree.ID,
		Path:            worktree.Path,
		AgentTmuxID:     worktree.AgentTmuxID,
		DevServerTmuxID: sql.NullString{Valid: false},
		ExternalUrl:     worktree.ExternalUrl,
	})
	
	return err
}

func handleTasksAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) > 0 {
			// Get specific task
			taskID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			
			task, err := queries.GetTask(ctx, taskID)
			if err != nil {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			
			json.NewEncoder(w).Encode(task)
		} else {
			// List all tasks - for now return empty array as we primarily use project tasks
			json.NewEncoder(w).Encode([]db.Task{})
		}
		
	case "PUT":
		if len(pathParts) == 0 {
			http.Error(w, "Task ID required", http.StatusBadRequest)
			return
		}
		
		taskID, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}
		
		var updateReq struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		updatedTask, err := queries.UpdateTask(ctx, db.UpdateTaskParams{
			ID:          taskID,
			Title:       updateReq.Title,
			Description: updateReq.Description,
			Status:      updateReq.Status,
		})
		if err != nil {
			log.Printf("Failed to update task: %v", err)
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}
		
		json.NewEncoder(w).Encode(updatedTask)
		
	case "DELETE":
		if len(pathParts) == 0 {
			http.Error(w, "Task ID required", http.StatusBadRequest)
			return
		}
		
		taskID, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid task ID", http.StatusBadRequest)
			return
		}
		
		// Check if there are any active task executions for this task
		executions, err := queries.GetTaskExecutionsByTaskID(ctx, taskID)
		if err == nil && len(executions) > 0 {
			// Delete all task executions first
			for _, execution := range executions {
				err := deleteTaskExecutionWithCleanup(ctx, execution.ID)
				if err != nil {
					log.Printf("Warning: failed to cleanup task execution %d: %v", execution.ID, err)
				}
			}
		}
		
		// Delete the task
		err = queries.DeleteTask(ctx, taskID)
		if err != nil {
			log.Printf("Failed to delete task: %v", err)
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProjectTasksAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) < 2 {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	
	projectID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "GET":
		// Get tasks for project - for now return empty array as tasks aren't fully implemented
		json.NewEncoder(w).Encode([]Task{})
		
	case "POST":
		// Create a new task for this project
		var createReq struct {
			Title           string `json:"title"`
			Description     string `json:"description"`
			Status          string `json:"status"`
			BaseDirectoryId string `json:"baseDirectoryId"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		if createReq.BaseDirectoryId == "" {
			http.Error(w, "Base directory ID is required", http.StatusBadRequest)
			return
		}
		
		// Create the task (no worktree creation needed)
		dbTask, err := queries.CreateTask(ctx, db.CreateTaskParams{
			ProjectID:       projectID,
			BaseDirectoryID: createReq.BaseDirectoryId,
			Title:           createReq.Title,
			Description:     createReq.Description,
			Status:          createReq.Status,
		})
		if err != nil {
			log.Printf("Failed to create task: %v", err)
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}
		
		// Get the base directory info to include in response
		dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err != nil {
			http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
			return
		}
		
		var baseDirectory BaseDirectory
		for _, dir := range dbBaseDirs {
			if dir.BaseDirectoryID == createReq.BaseDirectoryId {
				baseDirectory = dbBaseDirectoryToBaseDirectory(dir)
				break
			}
		}
		
		// Convert to API model
		task := dbTaskToTask(dbTask, baseDirectory)
		json.NewEncoder(w).Encode(task)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProjectBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	if len(pathParts) < 2 {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}
	
	projectID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}
	
	// Check if we have a specific directory ID in the path (for DELETE operations)
	if len(pathParts) >= 3 {
		directoryIDParam := pathParts[2]
		handleSingleBaseDirectoryAPI(w, r, ctx, projectID, directoryIDParam)
		return
	}
	
	switch r.Method {
	case "GET":
		// Get base directories for project
		dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
		if err != nil {
			log.Printf("Failed to get base directories for project %d: %v", projectID, err)
			json.NewEncoder(w).Encode([]BaseDirectory{})
			return
		}
		
		baseDirs := make([]BaseDirectory, 0)
		for _, dbBaseDir := range dbBaseDirs {
			baseDirs = append(baseDirs, dbBaseDirectoryToBaseDirectory(dbBaseDir))
		}
		
		json.NewEncoder(w).Encode(baseDirs)
		
	case "POST":
		// Create a new base directory for this project
		var createReq struct {
			Path                      string `json:"path"`
			GitInitialized            bool   `json:"gitInitialized"`
			WorktreeSetupCommands     string `json:"worktreeSetupCommands"`
			WorktreeTeardownCommands  string `json:"worktreeTeardownCommands"`
			DevServerSetupCommands    string `json:"devServerSetupCommands"`
			DevServerTeardownCommands string `json:"devServerTeardownCommands"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// Generate a unique base directory ID
		baseDirectoryID := fmt.Sprintf("bd_%d_%d", projectID, time.Now().Unix())
		
		dbBaseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
			ProjectID:                 projectID,
			BaseDirectoryID:           baseDirectoryID,
			Path:                      createReq.Path,
			GitInitialized:            createReq.GitInitialized,
			WorktreeSetupCommands:     createReq.WorktreeSetupCommands,
			WorktreeTeardownCommands:  createReq.WorktreeTeardownCommands,
			DevServerSetupCommands:    createReq.DevServerSetupCommands,
			DevServerTeardownCommands: createReq.DevServerTeardownCommands,
		})
		if err != nil {
			http.Error(w, "Failed to create base directory", http.StatusInternalServerError)
			return
		}
		
		result := dbBaseDirectoryToBaseDirectory(dbBaseDir)
		json.NewEncoder(w).Encode(result)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func loadTasksForProject(ctx context.Context, projectID int64, baseDirs []BaseDirectory) ([]Task, error) {
	dbTasks, err := queries.GetTasksByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	
	// Create a map for fast base directory lookup
	baseDirMap := make(map[string]BaseDirectory)
	for _, dir := range baseDirs {
		baseDirMap[dir.BaseDirectoryId] = dir
	}
	
	var tasks []Task
	for _, dbTask := range dbTasks {
		baseDir, exists := baseDirMap[dbTask.BaseDirectoryID]
		if !exists {
			// Skip tasks with missing base directories
			continue
		}
		tasks = append(tasks, dbTaskToTask(dbTask, baseDir))
	}
	
	return tasks, nil
}

func handleSingleBaseDirectoryAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, projectID int64, directoryIDParam string) {
    switch r.Method {
    case "PUT":
        // Update an existing base directory identified by its external ID within a project
        // First, fetch the existing directory to get its internal numeric ID
        existing, err := queries.GetBaseDirectoryByProjectAndID(ctx, db.GetBaseDirectoryByProjectAndIDParams{
            ProjectID:       projectID,
            BaseDirectoryID: directoryIDParam,
        })
        if err != nil {
            http.Error(w, "Base directory not found", http.StatusNotFound)
            return
        }

        // Parse update payload
        var updateReq struct {
            Path                      string `json:"path"`
            GitInitialized            bool   `json:"gitInitialized"`
            WorktreeSetupCommands     string `json:"worktreeSetupCommands"`
            WorktreeTeardownCommands  string `json:"worktreeTeardownCommands"`
            DevServerSetupCommands    string `json:"devServerSetupCommands"`
            DevServerTeardownCommands string `json:"devServerTeardownCommands"`
        }
        if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        updated, err := queries.UpdateBaseDirectory(ctx, db.UpdateBaseDirectoryParams{
            ID:                        existing.ID,
            Path:                      updateReq.Path,
            GitInitialized:            updateReq.GitInitialized,
            WorktreeSetupCommands:     updateReq.WorktreeSetupCommands,
            WorktreeTeardownCommands:  updateReq.WorktreeTeardownCommands,
            DevServerSetupCommands:    updateReq.DevServerSetupCommands,
            DevServerTeardownCommands: updateReq.DevServerTeardownCommands,
        })
        if err != nil {
            log.Printf("Failed to update base directory %d: %v", existing.ID, err)
            http.Error(w, "Failed to update base directory", http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(dbBaseDirectoryToBaseDirectory(updated))

    case "DELETE":
        // Find the directory by base_directory_id and project_id
        dbBaseDirs, err := queries.GetBaseDirectoriesByProjectID(ctx, projectID)
        if err != nil {
            http.Error(w, "Failed to get base directories", http.StatusInternalServerError)
			return
		}
		
		var directoryToDelete *db.BaseDirectory
		for _, dir := range dbBaseDirs {
			if dir.BaseDirectoryID == directoryIDParam {
				directoryToDelete = &dir
				break
			}
		}
		
		if directoryToDelete == nil {
			http.Error(w, "Base directory not found", http.StatusNotFound)
			return
		}
		
		// Delete the directory
		err = queries.DeleteBaseDirectory(ctx, directoryToDelete.ID)
		if err != nil {
			log.Printf("Failed to delete base directory %d: %v", directoryToDelete.ID, err)
			http.Error(w, "Failed to delete base directory", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func deleteTaskExecutionWithCleanup(ctx context.Context, executionID int64) error {
	log.Printf("Starting cleanup for task execution %d", executionID)
	
	// Get task execution details with all related information
	execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get task execution details: %v", err)
	}
	
	// Get worktree details
	worktree, err := queries.GetWorktree(ctx, execution.WorktreeID)
	if err != nil {
		log.Printf("Warning: failed to get worktree details: %v", err)
		// Continue with cleanup even if we can't get worktree details
	} else {
		// Perform tmux session cleanup
		err = cleanupTmuxSessions(worktree)
		if err != nil {
			log.Printf("Warning: failed to cleanup tmux sessions: %v", err)
		}
		
		// Get task details to find project ID
		task, err := queries.GetTask(ctx, execution.TaskID)
		if err != nil {
			log.Printf("Warning: failed to get task details: %v", err)
		} else {
			// Get base directory for teardown commands
			baseDir, err := queries.GetBaseDirectoryByProjectAndID(ctx, db.GetBaseDirectoryByProjectAndIDParams{
				ProjectID:       task.ProjectID,
				BaseDirectoryID: execution.BaseDirectoryID,
			})
			if err != nil {
				log.Printf("Warning: failed to get base directory for teardown commands: %v", err)
			} else {
				// Run teardown commands
				err = runTeardownCommands(worktree, baseDir)
				if err != nil {
					log.Printf("Warning: failed to run teardown commands: %v", err)
				}
				
				// Cleanup filesystem
				err = cleanupWorktreeDirectory(worktree, baseDir)
				if err != nil {
					log.Printf("Warning: failed to cleanup worktree directory: %v", err)
				}
			}
		}
		
		// Delete worktree record from database
		err = queries.DeleteWorktree(ctx, worktree.ID)
		if err != nil {
			log.Printf("Warning: failed to delete worktree from database: %v", err)
		}
	}
	
	// Delete task execution record from database
	err = queries.DeleteTaskExecution(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to delete task execution from database: %v", err)
	}
	
	log.Printf("Successfully cleaned up task execution %d", executionID)
	return nil
}

func cleanupTmuxSessions(worktree db.Worktree) error {
	// Kill agent tmux session if it exists
	if worktree.AgentTmuxID.Valid && worktree.AgentTmuxID.String != "" {
		log.Printf("Killing agent tmux session: %s", worktree.AgentTmuxID.String)
		cmd := exec.Command("tmux", "kill-session", "-t", worktree.AgentTmuxID.String)
		err := cmd.Run()
		if err != nil {
			log.Printf("Warning: failed to kill agent tmux session %s: %v", worktree.AgentTmuxID.String, err)
		}
	}
	
	// Kill dev server tmux session if it exists
	if worktree.DevServerTmuxID.Valid && worktree.DevServerTmuxID.String != "" {
		log.Printf("Killing dev server tmux session: %s", worktree.DevServerTmuxID.String)
		cmd := exec.Command("tmux", "kill-session", "-t", worktree.DevServerTmuxID.String)
		err := cmd.Run()
		if err != nil {
			log.Printf("Warning: failed to kill dev server tmux session %s: %v", worktree.DevServerTmuxID.String, err)
		}
	}
	
	return nil
}

func runTeardownCommands(worktree db.Worktree, baseDir db.BaseDirectory) error {
	// Run dev server teardown commands if they exist
	if baseDir.DevServerTeardownCommands != "" {
		log.Printf("Running dev server teardown commands: %s", baseDir.DevServerTeardownCommands)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", worktree.Path, baseDir.DevServerTeardownCommands))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Warning: dev server teardown commands failed: %v, output: %s", err, string(output))
		}
	}
	
	// Run worktree teardown commands if they exist
	if baseDir.WorktreeTeardownCommands != "" {
		log.Printf("Running worktree teardown commands: %s", baseDir.WorktreeTeardownCommands)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", worktree.Path, baseDir.WorktreeTeardownCommands))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Warning: worktree teardown commands failed: %v, output: %s", err, string(output))
		}
	}
	
	return nil
}

func cleanupWorktreeDirectory(worktree db.Worktree, baseDir db.BaseDirectory) error {
	// If this was a git worktree, remove it from git first
	if baseDir.GitInitialized {
		log.Printf("Removing git worktree: %s", worktree.Path)
		
		// Change to base directory and remove the git worktree
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && git worktree remove --force %s", baseDir.Path, worktree.Path))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Warning: failed to remove git worktree: %v, output: %s", err, string(output))
		}
	}
	
	// Remove the worktree directory from filesystem
	log.Printf("Removing worktree directory: %s", worktree.Path)
	err := os.RemoveAll(worktree.Path)
	if err != nil {
		return fmt.Errorf("failed to remove worktree directory %s: %v", worktree.Path, err)
	}
	
	return nil
}
