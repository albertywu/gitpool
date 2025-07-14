---
layout: default
title: GitPool - Lightning-Fast Git Worktree Management
description: Skip the wait. GitPool maintains a pool of pre-initialized Git worktrees for instant checkouts in CI/CD and development workflows.
---

<link rel="stylesheet" href="{{ '/assets/css/animations.css' | relative_url }}">
<script src="{{ '/assets/js/main.js' | relative_url }}" defer></script>

<div class="hero-section">
  <div class="hero-content">
    <h1 class="hero-title">
      <span class="gradient-text">GitPool</span>
    </h1>
    <p class="hero-subtitle">Lightning-fast Git worktree management for CI/CD and development workflows</p>
    <p class="hero-description">Skip the wait. Get instant access to pre-initialized Git worktrees without the overhead of cloning or fetching.</p>
    
    <div class="cta-buttons">
      <a href="#quick-start" class="btn btn-primary">Get Started</a>
      <a href="https://github.com/albertywu/gitpool" class="btn btn-secondary">
        <svg class="github-icon" viewBox="0 0 16 16" width="16" height="16">
          <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
        </svg>
        View on GitHub
      </a>
    </div>
  </div>
</div>

<div class="features-section">
  <div class="container">
    <h2 class="section-title">Why Choose GitPool?</h2>
    <div class="features-grid">
      <div class="feature-card">
        <div class="feature-icon">⚡</div>
        <h3>Instant Checkouts</h3>
        <p>No waiting for git operations - worktrees are pre-fetched and ready to use immediately.</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🚀</div>
        <h3>Perfect for CI/CD</h3>
        <p>Dramatically speed up build and test pipelines with zero-wait Git operations.</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">💾</div>
        <h3>Resource Efficient</h3>
        <p>Worktrees share Git objects with the source repository, saving disk space.</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🔒</div>
        <h3>Safe Isolation</h3>
        <p>Each claimed worktree is independent and protected from concurrent updates.</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🔄</div>
        <h3>Auto Maintenance</h3>
        <p>Background daemon keeps worktrees fresh and maintains pool capacity automatically.</p>
      </div>
      <div class="feature-card">
        <div class="feature-icon">🎯</div>
        <h3>Branch Management</h3>
        <p>Unique branch names with validation and automatic cleanup when released.</p>
      </div>
    </div>
  </div>
</div>

<div class="stats-section">
  <div class="container">
    <div class="stats-grid">
      <div class="stat-item">
        <div class="stat-number">0s</div>
        <div class="stat-label">Clone Time</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">10x</div>
        <div class="stat-label">Faster CI</div>
      </div>
      <div class="stat-item">
        <div class="stat-number">∞</div>
        <div class="stat-label">Parallel Jobs</div>
      </div>
    </div>
  </div>
</div>

<div id="quick-start" class="quick-start-section">
  <div class="container">
    <h2 class="section-title">Quick Start</h2>
    <div class="steps-grid">
      <div class="step">
        <div class="step-number">1</div>
        <h3>Install GitPool</h3>
        <div class="code-block">
          <code>go install github.com/albertywu/gitpool/cmd@latest</code>
        </div>
      </div>
      <div class="step">
        <div class="step-number">2</div>
        <h3>Start the Daemon</h3>
        <div class="code-block">
          <code>gitpool start</code>
        </div>
      </div>
      <div class="step">
        <div class="step-number">3</div>
        <h3>Add a Repository</h3>
        <div class="code-block">
          <code>gitpool track my-app ~/repos/my-app --max 8</code>
        </div>
      </div>
      <div class="step">
        <div class="step-number">4</div>
        <h3>Claim a Worktree</h3>
        <div class="code-block">
          <code>gitpool claim my-app --branch feature-xyz</code>
        </div>
      </div>
    </div>
  </div>
</div>

<div class="demo-section">
  <div class="container">
    <h2 class="section-title">See GitPool in Action</h2>
    
    <div class="terminal-container">
      <div class="terminal-demo" id="main-terminal"></div>
    </div>
    <div class="demo-comparison">
      <div class="demo-column">
        <h3 class="demo-title traditional">Traditional Git Workflow</h3>
        <div class="demo-steps">
          <div class="demo-step slow">
            <span class="step-icon">⏳</span>
            <span class="step-text">git clone (30s+)</span>
          </div>
          <div class="demo-step slow">
            <span class="step-icon">⏳</span>
            <span class="step-text">git fetch (10s+)</span>
          </div>
          <div class="demo-step slow">
            <span class="step-icon">⏳</span>
            <span class="step-text">git checkout (5s+)</span>
          </div>
          <div class="demo-step">
            <span class="step-icon">✅</span>
            <span class="step-text">Ready to work</span>
          </div>
        </div>
        <div class="demo-time slow-time">Total: ~45+ seconds</div>
      </div>
      
      <div class="demo-vs">VS</div>
      
      <div class="demo-column">
        <h3 class="demo-title gitpool">GitPool Workflow</h3>
        <div class="demo-steps">
          <div class="demo-step fast">
            <span class="step-icon">⚡</span>
            <span class="step-text">gitpool claim (&lt;1s)</span>
          </div>
          <div class="demo-step">
            <span class="step-icon">✅</span>
            <span class="step-text">Ready to work</span>
          </div>
        </div>
        <div class="demo-time fast-time">Total: &lt;1 second</div>
      </div>
    </div>
  </div>
</div>

<div class="use-cases-section">
  <div class="container">
    <h2 class="section-title">Perfect For</h2>
    <div class="use-cases-grid">
      <div class="use-case">
        <h3>🏗️ CI/CD Pipelines</h3>
        <p>Eliminate git clone overhead in your build and test pipelines. Get instant access to clean worktrees for every job.</p>
      </div>
      <div class="use-case">
        <h3>🧪 Parallel Testing</h3>
        <p>Run multiple test suites simultaneously across isolated worktrees without conflicts or waiting for checkouts.</p>
      </div>
      <div class="use-case">
        <h3>🔬 Quick Experiments</h3>
        <p>Instantly spin up isolated environments for trying out ideas without affecting your main worktree.</p>
      </div>
      <div class="use-case">
        <h3>🚢 Release Automation</h3>
        <p>Build releases from clean states without worrying about local modifications or stale dependencies.</p>
      </div>
    </div>
  </div>
</div>

<div class="cta-section">
  <div class="container">
    <h2>Ready to Speed Up Your Git Workflow?</h2>
    <p>Join developers who've eliminated git wait times from their CI/CD pipelines.</p>
    <div class="cta-buttons">
      <a href="/quickstart" class="btn btn-primary">Get Started Now</a>
      <a href="/docs" class="btn btn-secondary">Read the Docs</a>
    </div>
  </div>
</div>

<style>
/* Hero Section */
.hero-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 80px 20px;
  text-align: center;
  margin: -20px -20px 40px -20px;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 4rem;
  font-weight: 700;
  margin-bottom: 1rem;
  text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
}

.gradient-text {
  background: linear-gradient(45deg, #fff, #f0f8ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.hero-subtitle {
  font-size: 1.5rem;
  margin-bottom: 1rem;
  opacity: 0.9;
}

.hero-description {
  font-size: 1.1rem;
  margin-bottom: 2rem;
  opacity: 0.8;
  line-height: 1.6;
}

.cta-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 12px 24px;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 600;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.btn-primary {
  background: #fff;
  color: #667eea;
  border-color: #fff;
}

.btn-primary:hover {
  background: #f8f9fa;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}

.btn-secondary {
  background: transparent;
  color: #fff;
  border-color: #fff;
}

.btn-secondary:hover {
  background: #fff;
  color: #667eea;
  transform: translateY(-2px);
}

.github-icon {
  fill: currentColor;
}

/* Container */
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

/* Sections */
.features-section,
.quick-start-section,
.demo-section,
.use-cases-section {
  padding: 60px 0;
}

.stats-section {
  background: #f8f9fa;
  padding: 40px 0;
  margin: 40px 0;
}

.cta-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 60px 20px;
  text-align: center;
  margin: 60px -20px -20px -20px;
}

.section-title {
  text-align: center;
  font-size: 2.5rem;
  margin-bottom: 3rem;
  color: #333;
}

.cta-section .section-title,
.cta-section h2 {
  color: white;
  margin-bottom: 1rem;
}

/* Features Grid */
.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
}

.feature-card {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  text-align: center;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 25px rgba(0,0,0,0.15);
}

.feature-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.feature-card h3 {
  color: #333;
  margin-bottom: 1rem;
  font-size: 1.3rem;
}

.feature-card p {
  color: #666;
  line-height: 1.6;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
  text-align: center;
}

.stat-item {
  padding: 1rem;
}

.stat-number {
  font-size: 3rem;
  font-weight: 700;
  color: #667eea;
  margin-bottom: 0.5rem;
}

.stat-label {
  color: #666;
  font-weight: 500;
}

/* Quick Start Steps */
.steps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
}

.step {
  text-align: center;
  padding: 1.5rem;
}

.step-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  background: #667eea;
  color: white;
  border-radius: 50%;
  font-weight: 700;
  font-size: 1.2rem;
  margin-bottom: 1rem;
}

.step h3 {
  margin-bottom: 1rem;
  color: #333;
}

.code-block {
  background: #1a1a1a;
  color: #f8f8f2;
  padding: 1rem;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9rem;
  overflow-x: auto;
}

/* Demo Section */
.demo-comparison {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 2rem;
  align-items: start;
}

.demo-column {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.demo-title {
  text-align: center;
  margin-bottom: 1.5rem;
  font-size: 1.3rem;
}

.demo-title.traditional {
  color: #dc3545;
}

.demo-title.gitpool {
  color: #28a745;
}

.demo-steps {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.demo-step {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 6px;
}

.demo-step.slow {
  background: #ffeaa7;
  border-left: 4px solid #fdcb6e;
}

.demo-step.fast {
  background: #d1f2eb;
  border-left: 4px solid #00b894;
}

.step-icon {
  font-size: 1.2rem;
}

.step-text {
  font-weight: 500;
}

.demo-time {
  text-align: center;
  font-weight: 700;
  font-size: 1.1rem;
  padding: 0.75rem;
  border-radius: 6px;
}

.slow-time {
  background: #fab1a0;
  color: #d63031;
}

.fast-time {
  background: #55a3ff;
  color: white;
}

.demo-vs {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: #667eea;
  height: 100%;
}

/* Use Cases */
.use-cases-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 2rem;
}

.use-case {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  border-left: 4px solid #667eea;
}

.use-case h3 {
  color: #333;
  margin-bottom: 1rem;
  font-size: 1.2rem;
}

.use-case p {
  color: #666;
  line-height: 1.6;
}

/* Responsive Design */
@media (max-width: 768px) {
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-subtitle {
    font-size: 1.2rem;
  }
  
  .cta-buttons {
    flex-direction: column;
    align-items: center;
  }
  
  .demo-comparison {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  
  .demo-vs {
    order: 2;
    height: auto;
    padding: 1rem 0;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .features-grid,
  .steps-grid,
  .use-cases-grid {
    grid-template-columns: 1fr;
  }
}
</style>