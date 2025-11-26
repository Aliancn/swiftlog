---
layout: default
title: Contributing
nav_order: 10
description: "How to contribute to SwiftLog documentation"
---

# Contributing to SwiftLog Documentation

Thank you for your interest in contributing to SwiftLog documentation!

## How to Contribute

### Reporting Issues

If you find errors or unclear information:

1. Open an [issue on GitHub](https://github.com/aliancn/swiftlog/issues)
2. Use the label `documentation`
3. Describe the problem and suggest improvements

### Suggesting Improvements

Have ideas for new documentation?

1. Check existing [issues](https://github.com/aliancn/swiftlog/issues)
2. Open a new issue with the `enhancement` label
3. Describe what should be added or changed

### Submitting Changes

1. **Fork the repository**
   ```bash
   git clone https://github.com/your-username/swiftlog.git
   cd swiftlog
   ```

2. **Create a branch**
   ```bash
   git checkout -b docs/improve-xyz
   ```

3. **Make your changes**
   - Edit files in the `docs/` directory
   - Follow the style guide below
   - Test locally (see [LOCAL_TESTING](LOCAL_TESTING))

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "docs: improve XYZ documentation"
   ```

5. **Push and create PR**
   ```bash
   git push origin docs/improve-xyz
   ```
   Then open a Pull Request on GitHub.

## Documentation Style Guide

### File Format

All documentation files use Markdown with YAML front matter:

```markdown
---
title: Page Title
nav_order: 5
description: "Brief description"
---

# Page Title

Content here...
```

### Writing Style

- **Clear and Concise**: Use simple language
- **Active Voice**: "Run the command" not "The command should be run"
- **Present Tense**: "SwiftLog provides" not "SwiftLog will provide"
- **Code Examples**: Include working examples with output
- **Links**: Use relative links for internal pages

### Code Blocks

Use fenced code blocks with language specification:

````markdown
```bash
swiftlog run -- echo "Hello"
```

```javascript
const ws = new WebSocket('ws://localhost:8081');
```
````

### Headings

- Use `#` for page title (already in front matter)
- Use `##` for main sections
- Use `###` for subsections
- Use `####` for sub-subsections (avoid going deeper)

### Lists

**Unordered:**
```markdown
- Item one
- Item two
  - Nested item
```

**Ordered:**
```markdown
1. First step
2. Second step
3. Third step
```

### Tables

```markdown
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
```

### Callouts

Use Just the Docs callouts:

```markdown
{: .note }
This is a note.

{: .warning }
This is a warning.

{: .important }
This is important.
```

### Links

**Internal (relative):**
```markdown
See [Getting Started](getting-started)
See [Installation section](getting-started#installation)
```

**External:**
```markdown
Visit [GitHub](https://github.com)
```

## File Organization

```
docs/
├── index.md              # Home page (nav_order: 1)
├── getting-started.md    # Quick start (nav_order: 2)
├── architecture.md       # Architecture (nav_order: 3)
├── cli-guide.md          # CLI guide (nav_order: 4)
├── api-reference.md      # API docs (nav_order: 5)
├── configuration.md      # Config (nav_order: 6)
├── deployment.md         # Deployment (nav_order: 7)
├── development.md        # Development (nav_order: 8)
├── testing.md            # Testing (nav_order: 9)
└── CONTRIBUTING.md       # This file (nav_order: 10)
```

## Navigation Order

Pages appear in the sidebar based on `nav_order`:
- Lower numbers appear first
- Use increments of 1 for simplicity
- Reserve gaps for future additions

## Testing Your Changes

Before submitting:

1. **Test locally:**
   ```bash
   cd docs
   bundle install
   bundle exec jekyll serve
   ```

2. **Check for:**
   - Broken links
   - Formatting issues
   - Code example accuracy
   - Spelling and grammar

3. **Verify:**
   - Navigation order is correct
   - Search works properly
   - Mobile display is good

## Review Process

1. Submit PR with clear description
2. Maintainers will review
3. Address any feedback
4. Changes will be merged
5. Documentation site auto-updates

## Questions?

- Open an [issue](https://github.com/aliancn/swiftlog/issues)
- Join [discussions](https://github.com/aliancn/swiftlog/discussions)
- Contact maintainers

Thank you for contributing! 🎉
