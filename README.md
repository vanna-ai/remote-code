# Remote-Code

> ⚠️ **EXPERIMENTAL SOFTWARE** - This project is in active development. Expect breaking changes on each commit. No backwards compatibility will be supported yet.

Remote AI agent orchestration platform: Manage multiple coding agents, projects, and tasks from anywhere via web terminals, secure tunnels, and mobile app.

## What Remote-Code Does

Remote-Code is a web-based development environment that lets you **manage multiple AI coding agents from anywhere**. Unlike traditional development setups that tie you to a single machine, Remote-Code gives you a centralized command center that's accessible from any device with a browser.

### 🎛️ **Manage** - Complete AI Agent Orchestration

**Multi-Agent Coordination**
- Run multiple AI coding agents (Claude, Aider, Gemini, etc.) simultaneously
- Each agent operates in isolated tmux sessions to prevent conflicts
- Automatic agent detection and configuration
- Real-time monitoring of all active agents through web terminals

**Project & Task Organization**
- Organize development work across multiple projects and repositories
- Create and assign specific tasks to different agents
- Track task progress and execution history from a unified dashboard
- Configure base directories with custom setup/teardown automation

**Automated Workflow Management**
- Automatic git worktree creation for isolated development environments
- Integrated dev server management per project
- Custom setup and teardown commands for different project types
- Session persistence across disconnections

**Real-Time Control**
- Live terminal access to all running agents via WebSocket connections
- Send commands or guidance to agents mid-execution
- Monitor output and progress in real-time
- Complete visibility into what each agent is working on

### 🌍 **Anywhere** - True Location Independence

**Cross-Device Access**
- Full functionality through any modern web browser
- Works on desktop, laptop, tablet, and mobile devices
- Native mobile app in development for enhanced mobile experience
- No client software installation required

**Secure Remote Access**
- Cloudflare tunnel integration for secure public access
- ngrok support for quick URL generation
- Self-hosted architecture - your data stays on your infrastructure
- Access your development environment from coffee shops, airports, or home

**Persistent Sessions**
- tmux-powered sessions continue running even when you disconnect
- Resume work from any device exactly where you left off
- Background task execution - agents keep working while you're away
- Real-time notifications when tasks complete or need attention

## How It Works

1. **Set up projects** with base directories and configuration
2. **Configure AI agents** (Claude Code, Aider, etc.) through the web interface
3. **Create tasks** and assign them to specific agents
4. **Launch agents** in isolated tmux sessions with dedicated git worktrees
5. **Monitor and control** everything through the web dashboard
6. **Access from anywhere** using secure tunnel connections

## Key Features

- **Web-based terminals** with full terminal functionality via WebSocket
- **tmux session management** for persistent, isolated agent environments  
- **Git worktree automation** for conflict-free parallel development
- **Project lifecycle management** from setup to cleanup
- **Real-time dashboard** showing all projects, tasks, and agent status
- **Secure tunnel integration** for remote access (Cloudflare, ngrok)
- **Mobile-responsive interface** with native app in development
- **Agent communication** - send instructions to running agents
- **Background processing** - agents work autonomously

## Technology Stack

- **Backend**: Go with WebSocket support, SQLite database
- **Frontend**: SvelteKit with Tailwind CSS
- **Terminal**: tmux integration with WebSocket proxying
- **Database**: SQLite with comprehensive project/task/agent modeling
- **Remote Access**: Cloudflare tunnels, ngrok integration
- **Mobile**: Native app in development

## Getting Started

```bash
# Install dependencies
make install

# Build the application
make build

# Run the server
make run
```

The server starts on `http://localhost:8080` by default.

## Use Cases

- **Parallel AI Development**: Run multiple agents on different features simultaneously
- **Remote Code Reviews**: Monitor and guide AI agents from your phone or tablet  
- **Distributed Team Coordination**: Share agent sessions and project access
- **Long-running Tasks**: Start complex refactors, monitor progress remotely
- **Mobile Development Oversight**: Check build status and agent progress on the go
- **Cross-location Continuity**: Start work at the office, continue from home

## What Makes Remote-Code Different

Remote-Code combines the orchestration power of dedicated agent management tools with the accessibility of web-based development environments. Instead of being limited to a single machine or requiring complex SSH setups, you get a comprehensive platform that scales from solo development to team coordination - all accessible from any device, anywhere.

The focus is on **orchestration at scale** and **access from anywhere** - two capabilities that become essential when AI agents handle more of your development workload.

---

*Built for the era of AI-assisted development where managing multiple agents remotely is as important as the code they write.*