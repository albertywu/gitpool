// GitPool Website Interactive Features

document.addEventListener('DOMContentLoaded', function() {
    // Initialize all interactive features
    initScrollAnimations();
    initCopyCodeButtons();
    initTypewriter();
    initParallaxScroll();
    initSmoothScrolling();
    initAnimatedCounters();
    initTerminalSimulation();
});

// Scroll-based animations
function initScrollAnimations() {
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animate-in');
            }
        });
    }, observerOptions);

    // Observe elements for animation
    document.querySelectorAll('.feature-card, .step, .use-case, .demo-column').forEach(el => {
        observer.observe(el);
    });
}

// Copy code buttons
function initCopyCodeButtons() {
    document.querySelectorAll('pre code').forEach(codeBlock => {
        const button = document.createElement('button');
        button.className = 'copy-btn';
        button.innerHTML = '📋 Copy';
        button.addEventListener('click', () => copyCode(codeBlock, button));
        
        const pre = codeBlock.parentElement;
        pre.style.position = 'relative';
        pre.appendChild(button);
    });
}

function copyCode(codeBlock, button) {
    const text = codeBlock.textContent;
    navigator.clipboard.writeText(text).then(() => {
        button.innerHTML = '✅ Copied!';
        button.classList.add('copied');
        setTimeout(() => {
            button.innerHTML = '📋 Copy';
            button.classList.remove('copied');
        }, 2000);
    });
}

// Typewriter effect for hero title
function initTypewriter() {
    const title = document.querySelector('.hero-title .gradient-text');
    if (!title) return;
    
    const text = 'GitPool';
    title.textContent = '';
    
    let i = 0;
    const typeInterval = setInterval(() => {
        title.textContent += text[i];
        i++;
        if (i >= text.length) {
            clearInterval(typeInterval);
            title.classList.add('typewriter-complete');
        }
    }, 150);
}

// Parallax scrolling effect
function initParallaxScroll() {
    window.addEventListener('scroll', () => {
        const scrolled = window.pageYOffset;
        const parallaxElements = document.querySelectorAll('.hero-section');
        
        parallaxElements.forEach(element => {
            const speed = element.dataset.speed || 0.5;
            element.style.transform = `translateY(${scrolled * speed}px)`;
        });
    });
}

// Smooth scrolling for anchor links
function initSmoothScrolling() {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
}

// Animated counters
function initAnimatedCounters() {
    const counters = document.querySelectorAll('.stat-number');
    
    const observerOptions = {
        threshold: 0.5
    };
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                animateCounter(entry.target);
                observer.unobserve(entry.target);
            }
        });
    }, observerOptions);
    
    counters.forEach(counter => observer.observe(counter));
}

function animateCounter(element) {
    const text = element.textContent;
    const isNumber = /^\d+/.test(text);
    
    if (isNumber) {
        const finalNumber = parseInt(text);
        let current = 0;
        const increment = finalNumber / 50;
        const timer = setInterval(() => {
            current += increment;
            if (current >= finalNumber) {
                element.textContent = text; // Restore original text
                clearInterval(timer);
            } else {
                element.textContent = Math.floor(current) + text.replace(/^\d+/, '');
            }
        }, 50);
    } else {
        // For non-numeric counters like "0s", "10x", "∞"
        element.classList.add('counter-bounce');
    }
}

// Terminal simulation
function initTerminalSimulation() {
    const terminals = document.querySelectorAll('.terminal-demo');
    
    terminals.forEach(terminal => {
        simulateTerminal(terminal);
    });
}

function simulateTerminal(terminal) {
    const commands = [
        { cmd: '$ gitpool start', delay: 1000 },
        { output: '🚀 GitPool daemon started', delay: 500 },
        { cmd: '$ gitpool add my-app ~/repos/my-app --max 8', delay: 1500 },
        { output: '✅ Repository added to pool', delay: 500 },
        { cmd: '$ gitpool claim my-app --branch feature-xyz', delay: 1500 },
        { output: 'a91b6fc1-1234-5678-90ab-cdef12345678', delay: 300 },
        { output: '/home/user/.gitpool/worktrees/my-app/a91b6fc1-1234-5678-90ab-cdef12345678', delay: 300 },
        { cmd: '$ cd $WORKTREE_PATH && echo "Ready to work! ⚡"', delay: 1500 },
        { output: 'Ready to work! ⚡', delay: 500 }
    ];
    
    let index = 0;
    
    function typeCommand() {
        if (index >= commands.length) {
            setTimeout(() => {
                terminal.innerHTML = '';
                index = 0;
                typeCommand();
            }, 3000);
            return;
        }
        
        const item = commands[index];
        const line = document.createElement('div');
        line.className = item.cmd ? 'terminal-command' : 'terminal-output';
        terminal.appendChild(line);
        
        const text = item.cmd || item.output;
        let charIndex = 0;
        
        const typeInterval = setInterval(() => {
            line.textContent += text[charIndex];
            charIndex++;
            
            if (charIndex >= text.length) {
                clearInterval(typeInterval);
                index++;
                setTimeout(typeCommand, item.delay);
            }
        }, 50);
    }
    
    setTimeout(typeCommand, 1000);
}

// Add floating particles background
function initParticles() {
    const canvas = document.createElement('canvas');
    canvas.id = 'particles-canvas';
    canvas.style.position = 'fixed';
    canvas.style.top = '0';
    canvas.style.left = '0';
    canvas.style.width = '100%';
    canvas.style.height = '100%';
    canvas.style.pointerEvents = 'none';
    canvas.style.zIndex = '-1';
    canvas.style.opacity = '0.3';
    
    document.body.appendChild(canvas);
    
    const ctx = canvas.getContext('2d');
    let particles = [];
    
    function resizeCanvas() {
        canvas.width = window.innerWidth;
        canvas.height = window.innerHeight;
    }
    
    function createParticle() {
        return {
            x: Math.random() * canvas.width,
            y: Math.random() * canvas.height,
            vx: (Math.random() - 0.5) * 2,
            vy: (Math.random() - 0.5) * 2,
            size: Math.random() * 3 + 1,
            opacity: Math.random() * 0.5 + 0.2
        };
    }
    
    function initParticleSystem() {
        particles = [];
        for (let i = 0; i < 50; i++) {
            particles.push(createParticle());
        }
    }
    
    function updateParticles() {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        particles.forEach(particle => {
            particle.x += particle.vx;
            particle.y += particle.vy;
            
            if (particle.x < 0 || particle.x > canvas.width) particle.vx *= -1;
            if (particle.y < 0 || particle.y > canvas.height) particle.vy *= -1;
            
            ctx.beginPath();
            ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
            ctx.fillStyle = `rgba(102, 126, 234, ${particle.opacity})`;
            ctx.fill();
        });
        
        requestAnimationFrame(updateParticles);
    }
    
    window.addEventListener('resize', resizeCanvas);
    resizeCanvas();
    initParticleSystem();
    updateParticles();
}

// Initialize particles on hero section only
if (document.querySelector('.hero-section')) {
    initParticles();
}

// Add scroll progress indicator
function initScrollProgress() {
    const progressBar = document.createElement('div');
    progressBar.className = 'scroll-progress';
    document.body.appendChild(progressBar);
    
    window.addEventListener('scroll', () => {
        const scrollTop = window.pageYOffset;
        const docHeight = document.documentElement.scrollHeight - window.innerHeight;
        const scrollPercent = (scrollTop / docHeight) * 100;
        progressBar.style.width = scrollPercent + '%';
    });
}

initScrollProgress();

// Add theme toggle (dark/light mode)
function initThemeToggle() {
    const toggleButton = document.createElement('button');
    toggleButton.className = 'theme-toggle';
    toggleButton.innerHTML = '🌙';
    toggleButton.title = 'Toggle dark mode';
    
    document.body.appendChild(toggleButton);
    
    toggleButton.addEventListener('click', () => {
        document.body.classList.toggle('dark-theme');
        toggleButton.innerHTML = document.body.classList.contains('dark-theme') ? '☀️' : '🌙';
        
        // Save preference
        localStorage.setItem('darkMode', document.body.classList.contains('dark-theme'));
    });
    
    // Load saved preference
    if (localStorage.getItem('darkMode') === 'true') {
        document.body.classList.add('dark-theme');
        toggleButton.innerHTML = '☀️';
    }
}

initThemeToggle();