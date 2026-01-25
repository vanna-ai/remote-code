package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"remote-code/db"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robert-nix/ansihtml"
)

// -----------------
// Git API utilities
// -----------------

type GitFile struct {
	Path   string `json:"path"`
	X      string `json:"x"` // Index status
	Y      string `json:"y"` // Worktree status
	Staged bool   `json:"staged"`
}

type GitStatus struct {
	IsRepo         bool      `json:"isRepo"`
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

	// Files via porcelain -z, with -u to show individual files in untracked directories
	porcelain, _, err := runGit(dir, "status", "--porcelain=1", "-z", "-u")
	if err != nil {
		return nil, 1, porcelain, err
	}

	status := &GitStatus{IsRepo: true, StagedFiles: []GitFile{}, UnstagedFiles: []GitFile{}, UntrackedFiles: []GitFile{}, MergeConflicts: []GitFile{}}

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

func handleGitAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
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
		// If not a git repo, return a benign 200 with isRepo=false to avoid noisy 400s in UI.
		if out, _, err := runGit(dir, "rev-parse", "--is-inside-work-tree"); err != nil || strings.TrimSpace(out) != "true" {
			json.NewEncoder(w).Encode(&GitStatus{IsRepo: false, StagedFiles: []GitFile{}, UnstagedFiles: []GitFile{}, UntrackedFiles: []GitFile{}, MergeConflicts: []GitFile{}})
			return
		}
		st, _, _, err := getGitStatus(dir)
		if err != nil && st == nil {
			// Unexpected git error; still avoid 400 to keep UI clean
			json.NewEncoder(w).Encode(&GitStatus{IsRepo: false, StagedFiles: []GitFile{}, UnstagedFiles: []GitFile{}, UntrackedFiles: []GitFile{}, MergeConflicts: []GitFile{}})
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
		untracked := r.URL.Query().Get("untracked") == "true"

		var args []string
		if untracked && file != "" {
			// For untracked files, use --no-index to compare against /dev/null
			// This shows the entire file as additions
			args = []string{"diff", "--no-index", "/dev/null", file}
		} else {
			args = []string{"diff"}
			if staged {
				args = append(args, "--staged")
			}
			if file != "" {
				args = append(args, "--", file)
			}
		}

		out, code, err := runGit(dir, args...)
		// For --no-index, exit code 1 means differences found (which is expected)
		if err != nil && !(untracked && code == 1) {
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
		var body struct {
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
		var body struct {
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
		var body struct {
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
		var body struct {
			Path    string `json:"path"`
			Branch  string `json:"branch"`
			TaskID  int64  `json:"taskId"`
			AgentID int64  `json:"agentId"` // Agent ID of the winning execution for ELO
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid JSON: requires branch"})
			return
		}
		gitDir := body.Path
		if gitDir == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
			return
		}

		// Resolve base repository path
		commonDirOut, _, err := runGit(gitDir, "rev-parse", "--git-common-dir")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Not a git repo"})
			return
		}
		resolvedGitDir := strings.TrimSpace(commonDirOut)
		basePath := filepath.Dir(resolvedGitDir)

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

		// Record ELO competition for the merged execution (winner) vs all other executions
		eloStatus := "skipped"
		if body.TaskID != 0 && body.AgentID != 0 {
			// Process ELO competitions with the specified agent as winner
			eloCalc := NewELOCalculator(queries)
			if result, err := eloCalc.ProcessTaskCompetitionsWithWinner(ctx, body.TaskID, body.AgentID); err == nil {
				if result.TotalCompetitions > 0 {
					eloStatus = fmt.Sprintf("recorded %d competitions", result.TotalCompetitions)
				} else {
					eloStatus = "no competitions created"
				}
			} else {
				eloStatus = fmt.Sprintf("elo error: %v", err)
			}
		}

		// Optionally update task status to done
		taskStatus := "skipped"
		if body.TaskID != 0 {
			if task, err := queries.GetTask(ctx, body.TaskID); err == nil {
				if _, err := queries.UpdateTask(ctx, db.UpdateTaskParams{ID: task.ID, Title: task.Title, Description: task.Description, Status: "done"}); err == nil {
					taskStatus = "updated"
				} else {
					taskStatus = "failed"
				}
			} else {
				taskStatus = "failed"
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":         true,
			"output":     mergeOut,
			"taskStatus": taskStatus,
			"eloStatus":  eloStatus,
		})

	case "push":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
			return
		}
		var body struct {
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
				"ok":    false,
				"step":  "check_upstream",
				"error": "No upstream configured for main",
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

	case "merge-ready":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
			return
		}
		worktreeDir := r.URL.Query().Get("path")
		if worktreeDir == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing path"})
			return
		}
		// Worktree must be clean
		wtStatus, _, _ := runGit(worktreeDir, "status", "--porcelain")
		if strings.TrimSpace(wtStatus) != "" {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Uncommitted changes in worktree"})
			return
		}
		// Current worktree branch
		curBrOut, _, err := runGit(worktreeDir, "symbolic-ref", "--quiet", "--short", "HEAD")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Cannot detect current branch"})
			return
		}
		currentBranch := strings.TrimSpace(curBrOut)
		if currentBranch == "" || currentBranch == "main" {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Already on main"})
			return
		}
		// Resolve base repository path from any worktree dir via git-common-dir
		commonDirOut, _, err := runGit(worktreeDir, "rev-parse", "--git-common-dir")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Not a git repo"})
			return
		}
		gitDir := strings.TrimSpace(commonDirOut)
		basePath := filepath.Dir(gitDir)
		// Base must be clean
		baseStatus, _, _ := runGit(basePath, "status", "--porcelain")
		if strings.TrimSpace(baseStatus) != "" {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Base repo has uncommitted changes"})
			return
		}
		// SHAs for main and worktree HEAD
		headSHA, _, err := runGit(worktreeDir, "rev-parse", "HEAD")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Cannot resolve worktree HEAD"})
			return
		}
		branchSHA := strings.TrimSpace(headSHA)
		mainOut, _, err := runGit(basePath, "rev-parse", "main")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Cannot resolve main"})
			return
		}
		mainSHA := strings.TrimSpace(mainOut)
		if branchSHA == mainSHA {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "No changes to merge"})
			return
		}
		// Check fast-forward possibility: is main an ancestor of branch?
		// Run in the worktree to avoid any ambiguity with refs; we pass SHAs.
		_, code, _ := runGit(worktreeDir, "merge-base", "--is-ancestor", mainSHA, branchSHA)
		if code != 0 {
			json.NewEncoder(w).Encode(map[string]interface{}{"ready": false, "reason": "Non fast-forward; rebase required"})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"ready": true, "branch": currentBranch, "main": mainSHA, "head": branchSHA})

	case "update-from-main":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
			return
		}
		var body struct {
			Path     string `json:"path"`
			Strategy string `json:"strategy"` // "merge" (default) or "rebase"
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
		// Must be a git repo and clean
		wtStatus, _, _ := runGit(worktreeDir, "status", "--porcelain")
		if strings.TrimSpace(wtStatus) != "" {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "check_clean", "error": "Uncommitted changes in worktree"})
			return
		}
		// Determine current branch and ensure not main
		curBrOut, _, err := runGit(worktreeDir, "symbolic-ref", "--quiet", "--short", "HEAD")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "check_branch", "error": "Cannot detect current branch"})
			return
		}
		currentBranch := strings.TrimSpace(curBrOut)
		if currentBranch == "" || currentBranch == "main" {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "check_branch", "error": "Already on main"})
			return
		}
		// Fetch latest
		if out, code, err := runGit(worktreeDir, "fetch", "origin"); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "fetch", "code": code, "error": "git fetch failed", "output": out})
			return
		}
		strategy := strings.ToLower(strings.TrimSpace(body.Strategy))
		if strategy == "rebase" {
			out, code, err := runGit(worktreeDir, "rebase", "origin/main")
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "rebase", "code": code, "error": "git rebase failed", "output": out})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "output": out})
			return
		}
		// Default: merge origin/main into current branch (like GitHub Update branch)
		out, code, err := runGit(worktreeDir, "merge", "origin/main")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "step": "merge", "code": code, "error": "git merge failed", "output": out})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "output": out})

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
		var body struct {
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
	case "tasks":
		handleTasksAPI(w, r, ctx, pathParts[1:])
	case "task-executions":
		handleTaskExecutionsAPI(w, r, ctx, pathParts[1:])
	case "tmux-sessions":
		handleTmuxSessionsAPI(w, r, ctx, pathParts[1:])
	case "git":
		handleGitAPI(w, r, ctx, pathParts[1:])
	case "competitions":
		handleCompetitionsAPI(w, r, ctx, pathParts[1:])
	case "elo":
		handleELOAPI(w, r, ctx, pathParts[1:])
	default:
		http.Error(w, "Unknown API endpoint", http.StatusNotFound)
	}
}

type DashboardStats struct {
	ActiveSessions           int                      `json:"active_sessions"`
	Projects                 int                      `json:"projects"`
	TaskExecutions           int                      `json:"task_executions"`
	Agents                   int                      `json:"agents"`
	GitChangesAwaitingReview []TaskExecutionSummary   `json:"git_changes_awaiting_review"`
	AgentsWaitingForInput    []TaskExecutionSummary   `json:"agents_waiting_for_input"`
}

type TaskExecutionSummary struct {
	ID       int64  `json:"id"`
	TaskID   int64  `json:"task_id"`
	TaskName string `json:"task_name"`
	Agent    string `json:"agent"`
	Status   string `json:"status"`
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

			// Count task executions and populate special lists
			executions, err := queries.ListTaskExecutions(ctx)
			if err == nil {
				stats.TaskExecutions = len(executions)

				// Check each execution for agents waiting for input
				for _, execution := range executions {
					summary := TaskExecutionSummary{
						ID:       execution.ID,
						TaskID:   execution.TaskID,
						TaskName: execution.TaskTitle,
						Agent:    execution.AgentName,
						Status:   execution.Status,
					}

					// Skip rejected executions from dashboard sections
					if execution.Status == "rejected" {
						continue
					}

					// Check if agent is waiting for user input
					// First check stored status
					isWaiting := execution.Status == "Waiting" || execution.Status == "waiting"

					// If running, check real-time session status to detect waiting
					if !isWaiting && (execution.Status == "running" || execution.Status == "Running") {
						if execution.AgentTmuxID.Valid {
							realTimeStatus := determineTaskExecutionStatus(execution.AgentTmuxID.String)
							if realTimeStatus == "Waiting" {
								isWaiting = true
								summary.Status = "Waiting" // Update displayed status
							}
						}
					}

					if isWaiting {
						stats.AgentsWaitingForInput = append(stats.AgentsWaitingForInput, summary)
					}
				}
			}

			// Git changes are now tracked per base directory, not per task execution
			// Get all base directories and check their git status (reuse projects list from above)
			for _, project := range projects {
				baseDirs, bdErr := queries.GetBaseDirectoriesByProjectID(ctx, project.ID)
				if bdErr != nil {
					continue
				}
				for _, baseDir := range baseDirs {
					gitStatus, _, _, gitErr := getGitStatus(baseDir.Path)
					if gitErr == nil && gitStatus != nil && gitStatus.IsDirty {
						// Has uncommitted changes in this directory
						summary := TaskExecutionSummary{
							ID:       baseDir.ID,
							TaskID:   0,
							TaskName: baseDir.Path,
							Agent:    gitStatus.CurrentBranch,
							Status:   "dirty",
						}
						stats.GitChangesAwaitingReview = append(stats.GitChangesAwaitingReview, summary)
					}
				}
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
	Name        string  `json:"name"`
	Created     string  `json:"created"`
	Preview     string  `json:"preview"`
	TaskID      *int64  `json:"task_id,omitempty"`
	TaskName    *string `json:"task_name,omitempty"`
	AgentID     *int64  `json:"agent_id,omitempty"`
	AgentName   *string `json:"agent_name,omitempty"`
	ExecutionID *int64  `json:"execution_id,omitempty"`
	IsTask      bool    `json:"is_task"`
}

// SessionState stores the state of a tmux session for comparison
type SessionState struct {
	Name           string    `json:"name"`
	Content        string    `json:"content"`
	LastCursorPos  string    `json:"last_cursor_pos"`
	LastUpdated    time.Time `json:"last_updated"`
	UnchangedSince time.Time `json:"unchanged_since"`
	IsWaiting      bool      `json:"is_waiting"`
}

// Global storage for session states
var sessionStates = make(map[string]*SessionState)
var sessionStatesMutex sync.RWMutex

// Configuration for waiting detection
const WAITING_TIMEOUT = 30 * time.Second // Consider session waiting after 30 seconds of no change

// determineTaskExecutionStatus checks if a task execution is waiting based on its tmux session
func determineTaskExecutionStatus(sessionName string) string {
	if sessionName == "" {
		return "Running" // No session means not waiting
	}

	// Capture current session state
	currentState, err := captureSessionState(sessionName)
	if err != nil {
		// If we can't capture state, assume it's running (session might be ending, etc.)
		return "Running"
	}

	// Compare with previous state to determine if waiting
	status := compareSessionStates(sessionName, currentState)
	return status
}

// cleanupOrphanedSessionStates removes state entries for sessions that no longer exist
func cleanupOrphanedSessionStates() {
	// Get current tmux sessions to clean up orphaned states
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// If tmux isn't running or has no sessions, clear all states
		sessionStatesMutex.Lock()
		sessionStates = make(map[string]*SessionState)
		sessionStatesMutex.Unlock()
		return
	}

	currentSessionNames := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			currentSessionNames[line] = true
		}
	}

	// Remove states for sessions that no longer exist
	sessionStatesMutex.Lock()
	for sessionName := range sessionStates {
		if !currentSessionNames[sessionName] {
			delete(sessionStates, sessionName)
		}
	}
	sessionStatesMutex.Unlock()
}

// captureSessionState captures the current state of a tmux session for comparison
func captureSessionState(sessionName string) (*SessionState, error) {
	now := time.Now()

	// Capture the session content
	contentCmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-e", "-p")
	contentOutput, err := contentCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture session content: %v", err)
	}
	content := strings.TrimSpace(string(contentOutput))

	// Capture cursor position for more precise state detection
	cursorCmd := exec.Command("tmux", "display-message", "-t", sessionName, "-p", "#{cursor_x},#{cursor_y}")
	cursorOutput, err := cursorCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture cursor position: %v", err)
	}
	cursorPos := strings.TrimSpace(string(cursorOutput))

	return &SessionState{
		Name:           sessionName,
		Content:        content,
		LastCursorPos:  cursorPos,
		LastUpdated:    now,
		UnchangedSince: now, // Will be updated in compareSessionStates if unchanged
		IsWaiting:      false,
	}, nil
}

// compareSessionStates compares current state with previous state and updates waiting status
func compareSessionStates(sessionName string, currentState *SessionState) string {
	sessionStatesMutex.Lock()
	defer sessionStatesMutex.Unlock()

	previousState, exists := sessionStates[sessionName]

	// If no previous state exists, this is the first capture
	if !exists {
		sessionStates[sessionName] = currentState
		return "Running"
	}

	// Compare content and cursor position to detect changes
	hasChanged := previousState.Content != currentState.Content ||
		previousState.LastCursorPos != currentState.LastCursorPos

	if hasChanged {
		// Session has changed - reset waiting state
		currentState.UnchangedSince = currentState.LastUpdated
		currentState.IsWaiting = false
		sessionStates[sessionName] = currentState
		return "Running"
	} else {
		// Session hasn't changed - preserve the unchanged timestamp
		currentState.UnchangedSince = previousState.UnchangedSince

		// Check if it's been unchanged long enough to be considered waiting
		timeSinceLastChange := currentState.LastUpdated.Sub(currentState.UnchangedSince)
		if timeSinceLastChange >= WAITING_TIMEOUT {
			currentState.IsWaiting = true
			sessionStates[sessionName] = currentState
			return "Waiting"
		} else {
			currentState.IsWaiting = false
			sessionStates[sessionName] = currentState
			return "Running"
		}
	}
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

					// Look up execution ID for this task and agent combination
					if session.TaskID != nil {
						executions, err := queries.GetTaskExecutionsByTaskID(context.Background(), *session.TaskID)
						if err == nil {
							// Find the execution for this agent
							for _, exec := range executions {
								if exec.AgentID == agentID {
									session.ExecutionID = &exec.ID
									break
								}
							}
						}
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

	// Handle sub-endpoints like /api/task-executions/{id}/reject
	if len(pathParts) >= 2 && pathParts[1] == "reject" {
		handleRejectTaskExecution(w, r, ctx, pathParts)
		return
	}

	// Handle sub-endpoints like /api/task-executions/{id}/dev-server
	if len(pathParts) >= 2 && pathParts[1] == "dev-server" {
		executionID, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid execution ID", http.StatusBadRequest)
			return
		}

		queries := db.New(database)

		switch r.Method {
		case "POST":
			// Start dev server for task execution
			err = startDevServerForExecution(ctx, queries, executionID)
			if err != nil {
				log.Printf("Failed to start dev server: %v", err)
				http.Error(w, "Failed to start dev server", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server started"})
			return

		case "DELETE":
			// Stop dev server for task execution
			err = stopDevServerForExecution(ctx, queries, executionID)
			if err != nil {
				log.Printf("Failed to stop dev server: %v", err)
				http.Error(w, "Failed to stop dev server", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "dev server stopped"})
			return

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
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

			// Update status based on tmux session waiting detection if it has an active session
			if execution.AgentTmuxID.Valid {
				sessionName := execution.AgentTmuxID.String
				waitingStatus := determineTaskExecutionStatus(sessionName)
				if waitingStatus == "Waiting" {
					execution.Status = "Waiting"
				}
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

			// Clean up orphaned session states periodically
			cleanupOrphanedSessionStates()

			// Update status based on tmux session waiting detection
			for i := range executions {
				if executions[i].AgentTmuxID.Valid {
					sessionName := executions[i].AgentTmuxID.String
					waitingStatus := determineTaskExecutionStatus(sessionName)
					if waitingStatus == "Waiting" {
						executions[i].Status = "Waiting"
					}
				}
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

		// Clean up orphaned session states periodically
		cleanupOrphanedSessionStates()

		// Update status based on tmux session waiting detection
		for i := range executions {
			if executions[i].AgentTmuxID.Valid {
				sessionName := executions[i].AgentTmuxID.String
				waitingStatus := determineTaskExecutionStatus(sessionName)
				if waitingStatus == "Waiting" {
					executions[i].Status = "Waiting"
				}
			}
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

		// Get the base directory
		dbBaseDir, err := queries.GetBaseDirectoryByProjectAndID(ctx, db.GetBaseDirectoryByProjectAndIDParams{
			ProjectID:       dbTask.ProjectID,
			BaseDirectoryID: dbTask.BaseDirectoryID,
		})
		if err != nil {
			log.Printf("Failed to get base directory: %v", err)
			http.Error(w, "Base directory not found", http.StatusNotFound)
			return
		}

		// Create the task execution record (no worktree - runs directly in base directory)
		dbTaskExecution, err := queries.CreateTaskExecution(ctx, db.CreateTaskExecutionParams{
			TaskID:          createReq.TaskId,
			AgentID:         createReq.AgentId,
			Status:          "starting",
			AgentTmuxID:     sql.NullString{Valid: false},
			DevServerTmuxID: sql.NullString{Valid: false},
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
		go startTaskExecutionProcess(dbTaskExecution.ID, dbTask, dbAgent, dbBaseDir)

		// Return the task execution details
		result := map[string]interface{}{
			"id":                  dbTaskExecution.ID,
			"task_id":             dbTaskExecution.TaskID,
			"agent_id":            dbTaskExecution.AgentID,
			"base_directory_path": dbBaseDir.Path,
			"status":              dbTaskExecution.Status,
		}
		json.NewEncoder(w).Encode(result)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func startTaskExecutionProcess(executionID int64, task db.Task, agent db.Agent, baseDir db.BaseDirectory) {
	ctx := context.Background()

	log.Printf("Starting task execution %d: Task '%s' with agent '%s' in %s", executionID, task.Title, agent.Name, baseDir.Path)

	// Generate a unique tmux session name
	sessionName := fmt.Sprintf("task_%d_agent_%d", task.ID, agent.ID)

	// Start tmux session in the base directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", baseDir.Path)
	err := tmuxCmd.Run()
	if err != nil {
		log.Printf("Failed to start tmux session: %v", err)
		updateTaskExecutionStatus(ctx, executionID, "failed")
		return
	}

	// Update task execution with tmux session info
	_, err = queries.UpdateTaskExecutionTmux(ctx, db.UpdateTaskExecutionTmuxParams{
		ID:              executionID,
		AgentTmuxID:     sql.NullString{String: sessionName, Valid: true},
		DevServerTmuxID: sql.NullString{Valid: false},
	})
	if err != nil {
		log.Printf("Failed to update task execution with tmux session: %v", err)
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

	// Run custom setup commands if provided
	if baseDir.SetupCommands != "" {
		if err := sendCommandAndWait(baseDir.SetupCommands, "setup commands"); err != nil {
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
		return
	}

	// ELO competitions are now recorded when user merges, not on task completion
}

// processTaskCompetitionsAsync function removed - ELO competitions now recorded on merge, not task completion

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
	execution, err := queries.GetTaskExecution(ctx, executionID)
	if err != nil {
		log.Printf("Failed to get task execution: %v", err)
		http.Error(w, "Task execution not found", http.StatusNotFound)
		return
	}

	if !execution.AgentTmuxID.Valid {
		http.Error(w, "No active tmux session for this task execution", http.StatusBadRequest)
		return
	}

	sessionName := execution.AgentTmuxID.String

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
	execution, err := queries.GetTaskExecution(ctx, executionID)
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

	if !execution.AgentTmuxID.Valid {
		http.Error(w, "No active tmux session for this task execution", http.StatusBadRequest)
		return
	}

	sessionName := execution.AgentTmuxID.String

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
		"success":          true,
		"message":          "Task prompt re-sent to session successfully",
		"session":          sessionName,
		"task_title":       task.Title,
		"task_description": task.Description,
	}
	json.NewEncoder(w).Encode(response)
}

func handleRejectTaskExecution(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
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

	// Get the task execution to check current status
	execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
	if err != nil {
		log.Printf("Failed to get task execution: %v", err)
		http.Error(w, "Task execution not found", http.StatusNotFound)
		return
	}

	// Check if already rejected
	if execution.Status == "rejected" {
		http.Error(w, "Task execution is already rejected", http.StatusBadRequest)
		return
	}

	// Update status to "rejected"
	_, err = queries.UpdateTaskExecutionStatus(ctx, db.UpdateTaskExecutionStatusParams{
		ID:     executionID,
		Status: "rejected",
	})
	if err != nil {
		log.Printf("Failed to update task execution status: %v", err)
		http.Error(w, "Failed to reject task execution", http.StatusInternalServerError)
		return
	}

	// Record ELO losses against all other agents for the same task
	go func() {
		// Use a new context for background processing
		bgCtx := context.Background()
		bgQueries := db.New(database)
		
		err := recordELOLossesForRejectedTask(bgCtx, bgQueries, execution.TaskID, execution.AgentID)
		if err != nil {
			log.Printf("Failed to record ELO losses for rejected task %d: %v", execution.TaskID, err)
		}
	}()

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Task execution rejected successfully",
		"status":  "rejected",
	}
	json.NewEncoder(w).Encode(response)
}

func handleBaseDirectoriesAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			// List all base directories
			// For now, get them via projects
			projects, err := queries.ListProjects(ctx)
			if err != nil {
				http.Error(w, "Failed to list projects", http.StatusInternalServerError)
				return
			}

			var allDirs []map[string]interface{}
			for _, project := range projects {
				dirs, err := queries.GetBaseDirectoriesByProjectID(ctx, project.ID)
				if err != nil {
					continue
				}
				for _, dir := range dirs {
					allDirs = append(allDirs, map[string]interface{}{
						"id":                           dir.ID,
						"base_directory_id":            dir.BaseDirectoryID,
						"project_id":                   dir.ProjectID,
						"path":                         dir.Path,
						"git_initialized":              dir.GitInitialized,
						"setup_commands":               dir.SetupCommands,
						"teardown_commands":            dir.TeardownCommands,
						"dev_server_setup_commands":    dir.DevServerSetupCommands,
						"dev_server_teardown_commands": dir.DevServerTeardownCommands,
					})
				}
			}
			json.NewEncoder(w).Encode(allDirs)
			return
		}

		// Get single base directory by internal ID
		id, err := strconv.ParseInt(pathParts[0], 10, 64)
		if err != nil {
			http.Error(w, "Invalid directory ID", http.StatusBadRequest)
			return
		}

		dir, err := queries.GetBaseDirectory(ctx, id)
		if err != nil {
			http.Error(w, "Base directory not found", http.StatusNotFound)
			return
		}

		// Get the project name
		project, _ := queries.GetProject(ctx, dir.ProjectID)
		projectName := ""
		if project.Name != "" {
			projectName = project.Name
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                           dir.ID,
			"base_directory_id":            dir.BaseDirectoryID,
			"project_id":                   dir.ProjectID,
			"project_name":                 projectName,
			"path":                         dir.Path,
			"git_initialized":              dir.GitInitialized,
			"setup_commands":               dir.SetupCommands,
			"teardown_commands":            dir.TeardownCommands,
			"dev_server_setup_commands":    dir.DevServerSetupCommands,
			"dev_server_teardown_commands": dir.DevServerTeardownCommands,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Dev server functions for task executions
func startDevServerForExecution(ctx context.Context, queries *db.Queries, executionID int64) error {
	// Get task execution with details
	execution, err := queries.GetTaskExecutionWithDetails(ctx, executionID)
	if err != nil {
		return err
	}

	// Get task with base directory info for dev server setup commands
	taskWithBaseDir, err := queries.GetTaskWithBaseDirectory(ctx, execution.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task with base directory: %v", err)
	}

	// Create dev server session name
	devSessionName := fmt.Sprintf("dev_%d", executionID)

	// Start tmux session for dev server in the base directory
	tmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", devSessionName, "-c", execution.BaseDirectoryPath)
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

		// Add info message
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server started. Session: "+devSessionName+"'", "Enter")
		echoCmd.Run()
	} else {
		// No setup commands, just show info
		echoCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "echo 'Dev server session created. No setup commands configured.'", "Enter")
		echoCmd.Run()

		bashCmd := exec.Command("tmux", "send-keys", "-t", devSessionName, "bash", "Enter")
		bashCmd.Run()
	}

	// Update task execution with dev server session info
	_, err = queries.UpdateTaskExecutionTmux(ctx, db.UpdateTaskExecutionTmuxParams{
		ID:              executionID,
		AgentTmuxID:     execution.AgentTmuxID,
		DevServerTmuxID: sql.NullString{String: devSessionName, Valid: true},
	})

	return err
}

func stopDevServerForExecution(ctx context.Context, queries *db.Queries, executionID int64) error {
	// Get task execution
	execution, err := queries.GetTaskExecution(ctx, executionID)
	if err != nil {
		return err
	}

	if execution.DevServerTmuxID.Valid {
		// Kill the tmux session
		tmuxCmd := exec.Command("tmux", "kill-session", "-t", execution.DevServerTmuxID.String)
		_ = tmuxCmd.Run() // Ignore error if session doesn't exist
	}

	// Update task execution to remove dev server session info
	_, err = queries.UpdateTaskExecutionTmux(ctx, db.UpdateTaskExecutionTmuxParams{
		ID:              executionID,
		AgentTmuxID:     execution.AgentTmuxID,
		DevServerTmuxID: sql.NullString{Valid: false},
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
			SetupCommands             string `json:"setupCommands"`
			TeardownCommands          string `json:"teardownCommands"`
			WorktreeSetupCommands     string `json:"worktreeSetupCommands"`     // Legacy support
			WorktreeTeardownCommands  string `json:"worktreeTeardownCommands"`  // Legacy support
			DevServerSetupCommands    string `json:"devServerSetupCommands"`
			DevServerTeardownCommands string `json:"devServerTeardownCommands"`
		}

		if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Support legacy field names
		setupCommands := createReq.SetupCommands
		if setupCommands == "" {
			setupCommands = createReq.WorktreeSetupCommands
		}
		teardownCommands := createReq.TeardownCommands
		if teardownCommands == "" {
			teardownCommands = createReq.WorktreeTeardownCommands
		}

		// Generate a unique base directory ID
		baseDirectoryID := fmt.Sprintf("bd_%d_%d", projectID, time.Now().Unix())

		dbBaseDir, err := queries.CreateBaseDirectory(ctx, db.CreateBaseDirectoryParams{
			ProjectID:                 projectID,
			BaseDirectoryID:           baseDirectoryID,
			Path:                      createReq.Path,
			GitInitialized:            createReq.GitInitialized,
			SetupCommands:             setupCommands,
			TeardownCommands:          teardownCommands,
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
			SetupCommands             string `json:"setupCommands"`
			TeardownCommands          string `json:"teardownCommands"`
			WorktreeSetupCommands     string `json:"worktreeSetupCommands"`     // Legacy support
			WorktreeTeardownCommands  string `json:"worktreeTeardownCommands"`  // Legacy support
			DevServerSetupCommands    string `json:"devServerSetupCommands"`
			DevServerTeardownCommands string `json:"devServerTeardownCommands"`
		}
		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Support legacy field names
		setupCommands := updateReq.SetupCommands
		if setupCommands == "" {
			setupCommands = updateReq.WorktreeSetupCommands
		}
		teardownCommands := updateReq.TeardownCommands
		if teardownCommands == "" {
			teardownCommands = updateReq.WorktreeTeardownCommands
		}

		updated, err := queries.UpdateBaseDirectory(ctx, db.UpdateBaseDirectoryParams{
			ID:                        existing.ID,
			Path:                      updateReq.Path,
			GitInitialized:            updateReq.GitInitialized,
			SetupCommands:             setupCommands,
			TeardownCommands:          teardownCommands,
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

	// Perform tmux session cleanup using task execution's tmux IDs
	cleanupTmuxSessionsFromExecution(execution)

	// Get task details to find project ID for teardown commands
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
			// Run teardown commands in base directory
			runTeardownCommandsInDir(baseDir)
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

func cleanupTmuxSessionsFromExecution(execution db.GetTaskExecutionWithDetailsRow) {
	// Kill agent tmux session if it exists
	if execution.AgentTmuxID.Valid && execution.AgentTmuxID.String != "" {
		log.Printf("Killing agent tmux session: %s", execution.AgentTmuxID.String)
		cmd := exec.Command("tmux", "kill-session", "-t", execution.AgentTmuxID.String)
		err := cmd.Run()
		if err != nil {
			log.Printf("Warning: failed to kill agent tmux session %s: %v", execution.AgentTmuxID.String, err)
		}
	}

	// Kill dev server tmux session if it exists
	if execution.DevServerTmuxID.Valid && execution.DevServerTmuxID.String != "" {
		log.Printf("Killing dev server tmux session: %s", execution.DevServerTmuxID.String)
		cmd := exec.Command("tmux", "kill-session", "-t", execution.DevServerTmuxID.String)
		err := cmd.Run()
		if err != nil {
			log.Printf("Warning: failed to kill dev server tmux session %s: %v", execution.DevServerTmuxID.String, err)
		}
	}
}

func runTeardownCommandsInDir(baseDir db.BaseDirectory) {
	// Run dev server teardown commands if they exist
	if baseDir.DevServerTeardownCommands != "" {
		log.Printf("Running dev server teardown commands: %s", baseDir.DevServerTeardownCommands)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", baseDir.Path, baseDir.DevServerTeardownCommands))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Warning: dev server teardown commands failed: %v, output: %s", err, string(output))
		}
	}

	// Run teardown commands if they exist
	if baseDir.TeardownCommands != "" {
		log.Printf("Running teardown commands: %s", baseDir.TeardownCommands)
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", baseDir.Path, baseDir.TeardownCommands))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Warning: teardown commands failed: %v, output: %s", err, string(output))
		}
	}
}

func handleCompetitionsAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 {
			competitions, err := queries.ListCompetitions(ctx)
			if err != nil {
				http.Error(w, "Failed to list competitions", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(competitions)
		} else if pathParts[0] == "history" {
			history, err := queries.GetCompetitionHistory(ctx)
			if err != nil {
				http.Error(w, "Failed to get competition history", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(history)
		} else if pathParts[0] == "task" && len(pathParts) > 1 {
			taskID, err := strconv.ParseInt(pathParts[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}
			competitions, err := queries.ListCompetitionsByTask(ctx, taskID)
			if err != nil {
				http.Error(w, "Failed to get competitions for task", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(competitions)
		} else if pathParts[0] == "agent" && len(pathParts) > 1 {
			agentID, err := strconv.ParseInt(pathParts[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}
			competitions, err := queries.ListCompetitionsByAgent(ctx, db.ListCompetitionsByAgentParams{
				Agent1ID: agentID,
				Agent2ID: agentID,
			})
			if err != nil {
				http.Error(w, "Failed to get competitions for agent", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(competitions)
		} else {
			competitionID, err := strconv.ParseInt(pathParts[0], 10, 64)
			if err != nil {
				http.Error(w, "Invalid competition ID", http.StatusBadRequest)
				return
			}
			competition, err := queries.GetCompetition(ctx, competitionID)
			if err != nil {
				http.Error(w, "Competition not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(competition)
		}

	case "POST":
		if len(pathParts) > 0 && pathParts[0] == "process-task" && len(pathParts) > 1 {
			taskID, err := strconv.ParseInt(pathParts[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}

			eloCalc := NewELOCalculator(queries)
			result, err := eloCalc.ProcessTaskCompetitions(ctx, taskID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to process competitions: %v", err), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(result)
		} else {
			var createReq struct {
				TaskID            int64  `json:"task_id"`
				Agent1ID          int64  `json:"agent1_id"`
				Agent2ID          int64  `json:"agent2_id"`
				Agent1ExecutionID int64  `json:"agent1_execution_id"`
				Agent2ExecutionID int64  `json:"agent2_execution_id"`
				Result            string `json:"result"` // "agent1_wins", "agent2_wins", "draw"
				Notes             string `json:"notes"`
			}

			if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			var result MatchResult
			switch createReq.Result {
			case "agent1_wins":
				result = Agent1Wins
			case "agent2_wins":
				result = Agent2Wins
			case "draw":
				result = Draw
			default:
				http.Error(w, "Invalid result value", http.StatusBadRequest)
				return
			}

			eloCalc := NewELOCalculator(queries)
			competition, err := eloCalc.RecordCompetition(ctx, CompetitionParams{
				TaskID:            createReq.TaskID,
				Agent1ID:          createReq.Agent1ID,
				Agent2ID:          createReq.Agent2ID,
				Agent1ExecutionID: createReq.Agent1ExecutionID,
				Agent2ExecutionID: createReq.Agent2ExecutionID,
				Result:            result,
				Notes:             createReq.Notes,
			})

			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create competition: %v", err), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(competition)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleELOAPI(w http.ResponseWriter, r *http.Request, ctx context.Context, pathParts []string) {
	switch r.Method {
	case "GET":
		if len(pathParts) == 0 || pathParts[0] == "leaderboard" {
			leaderboard, err := queries.GetAgentLeaderboard(ctx)
			if err != nil {
				http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(leaderboard)
		} else if pathParts[0] == "agent" && len(pathParts) > 2 && pathParts[2] == "history" {
			agentID, err := strconv.ParseInt(pathParts[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid agent ID", http.StatusBadRequest)
				return
			}
			history, err := queries.GetAgentELOHistory(ctx, db.GetAgentELOHistoryParams{
				Agent1ID: agentID,
				Agent2ID: agentID,
			})
			if err != nil {
				http.Error(w, "Failed to get ELO history", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(history)
		} else if pathParts[0] == "head-to-head" && len(pathParts) > 2 {
			agent1ID, err := strconv.ParseInt(pathParts[1], 10, 64)
			if err != nil {
				http.Error(w, "Invalid agent1 ID", http.StatusBadRequest)
				return
			}
			agent2ID, err := strconv.ParseInt(pathParts[2], 10, 64)
			if err != nil {
				http.Error(w, "Invalid agent2 ID", http.StatusBadRequest)
				return
			}

			record, err := queries.GetHeadToHeadRecord(ctx, db.GetHeadToHeadRecordParams{
				WinnerAgentID:   sql.NullInt64{Valid: true, Int64: agent1ID},
				WinnerAgentID_2: sql.NullInt64{Valid: true, Int64: agent2ID},
				Agent1ID:        agent1ID,
				Agent2ID:        agent2ID,
				Agent1ID_2:      agent2ID,
				Agent2ID_2:      agent1ID,
			})
			if err != nil {
				http.Error(w, "Failed to get head-to-head record", http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(record)
		} else {
			http.Error(w, "Unknown ELO endpoint", http.StatusNotFound)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
