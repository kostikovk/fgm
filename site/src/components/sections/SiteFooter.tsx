export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="container footer-inner">
        <div>
          <strong>FGM (Fast Go Manager)</strong> · Open-source CLI for Go &amp;{" "}
          <code>golangci-lint</code> workflows.
        </div>
        <nav className="footer-links" aria-label="Footer">
          <a href="https://github.com/kostikovk/fgm/blob/main/docs/README.md">
            Docs
          </a>
          <a href="https://github.com/kostikovk/fgm/blob/main/docs/reference.md">
            CLI Reference
          </a>
          <a href="https://github.com/kostikovk/fgm/releases">Releases</a>
          <a href="https://github.com/kostikovk/fgm/issues">Issues</a>
          <a href="https://github.com/kostikovk/fgm/blob/main/LICENSE">
            MIT License
          </a>
        </nav>
      </div>
    </footer>
  );
}
