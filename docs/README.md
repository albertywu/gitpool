# GitPool GitHub Pages Website

This directory contains the source files for the GitPool GitHub Pages website at https://albertywu.github.io/gitpool.

## Structure

```
docs/
├── _config.yml           # Jekyll configuration
├── _layouts/             # Custom layouts
│   └── default.html      # Main layout template
├── assets/               # Static assets
│   ├── css/              # Stylesheets
│   │   └── animations.css
│   └── js/               # JavaScript files
│       └── main.js
├── index.md              # Homepage
├── quickstart.md         # Quick start guide
├── docs.md               # Documentation
├── examples.md           # Examples and use cases
└── README.md             # This file
```

## Features

- **Modern Design**: Clean, professional layout with gradient backgrounds and smooth animations
- **Interactive Elements**: 
  - Scroll-triggered animations
  - Copy-to-clipboard code blocks
  - Terminal simulation
  - Animated counters
  - Floating particles background
  - Dark/light theme toggle
- **Responsive**: Mobile-first design that works on all devices
- **Performance**: Optimized assets and minimal dependencies
- **SEO**: Proper meta tags, structured data, and semantic HTML

## Development

### Local Setup

1. Install Jekyll:
   ```bash
   gem install jekyll bundler
   ```

2. Navigate to docs directory:
   ```bash
   cd docs
   ```

3. Install dependencies:
   ```bash
   bundle install
   ```

4. Serve locally:
   ```bash
   bundle exec jekyll serve
   ```

5. Visit http://localhost:4000

### Making Changes

- Edit Markdown files for content changes
- Modify `assets/css/animations.css` for styling
- Update `assets/js/main.js` for interactive features
- Customize `_layouts/default.html` for layout changes

### Deployment

The site is automatically deployed via GitHub Actions when changes are pushed to the `main` branch in the `docs/` directory.

## Browser Support

- Chrome/Chromium 70+
- Firefox 65+
- Safari 12+
- Edge 79+

## Performance Features

- Lazy loading for animations
- Optimized asset delivery
- Minimal JavaScript bundle
- CSS-only animations where possible
- Reduced motion support for accessibility

## Accessibility

- Semantic HTML structure
- Proper heading hierarchy
- Alt text for images
- Keyboard navigation support
- High contrast ratios
- Reduced motion preferences respected