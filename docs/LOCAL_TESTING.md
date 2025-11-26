# Testing Documentation Locally

This guide shows how to test the documentation site locally before publishing to GitHub Pages.

## Prerequisites

- Ruby 2.7 or higher
- Bundler

### Install Ruby (if needed)

**macOS:**
```bash
brew install ruby
```

**Ubuntu/Debian:**
```bash
sudo apt-get install ruby-full build-essential
```

**Windows:**
Download from [RubyInstaller](https://rubyinstaller.org/)

## Setup

1. **Navigate to docs directory:**
```bash
cd docs
```

2. **Install Bundler:**
```bash
gem install bundler
```

3. **Install dependencies:**
```bash
bundle install
```

This will install Jekyll and all required gems.

## Running Locally

### Start the server:
```bash
bundle exec jekyll serve
```

Or with live reload:
```bash
bundle exec jekyll serve --livereload
```

### Access the site:
Open your browser to: **http://localhost:4000/swiftlog/**

The server will automatically rebuild when you make changes to files.

## Building the Site

To build the site without serving:
```bash
bundle exec jekyll build
```

The generated site will be in the `_site/` directory.

## Troubleshooting

### Bundler not found
```bash
gem install bundler
```

### Permission errors
```bash
# Install gems to user directory
bundle install --path vendor/bundle
```

### Port already in use
```bash
# Use a different port
bundle exec jekyll serve --port 4001
```

### Clean build
```bash
# Remove generated files and rebuild
bundle exec jekyll clean
bundle exec jekyll build
```

## Configuration

The site is configured via `_config.yml`:
- Change `baseurl` if testing with different paths
- Modify `url` for different domains
- Adjust theme settings as needed

## Theme Documentation

SwiftLog docs use the [Just the Docs](https://just-the-docs.github.io/just-the-docs/) theme.

For theme customization, see:
- [Just the Docs Documentation](https://just-the-docs.github.io/just-the-docs/)
- [Configuration Options](https://just-the-docs.github.io/just-the-docs/docs/configuration/)

## Publishing to GitHub Pages

Once you're happy with local testing:

1. Commit changes:
```bash
git add .
git commit -m "Update documentation"
git push
```

2. Enable GitHub Pages:
   - Go to repository Settings
   - Navigate to Pages section
   - Source: `main` branch, `/docs` folder
   - Save

The site will be published to: **https://aliancn.github.io/swiftlog/**

## Tips

- Use `--drafts` to preview draft posts: `bundle exec jekyll serve --drafts`
- Use `--incremental` for faster builds: `bundle exec jekyll serve --incremental`
- Check for errors: `bundle exec jekyll build --verbose`
