package ant

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/yields/ant/internal/scan"
	"github.com/yields/ant/internal/selectors"
	"golang.org/x/net/html"
)

// Page represents a page.
type Page struct {
	URL    *url.URL
	Header http.Header
	body   io.ReadCloser
	root   *html.Node
	once   sync.Once
	err    error
}

// Body returns the raw body of the page.
//
// Note: if the body is read, the page's methods will not be available.
func (p *Page) Body() io.Reader {
	return p.body
}

// Document returns the parsed document.
//
// The method returns an error if the document could not be parsed.
func (p *Page) Document() (*html.Node, error) {
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.root, nil
}

// HTML returns the parsed HTML document.
//
// The method returns an error if the document could not be parsed.
func (p *Page) HTML() (string, error) {
	if err := p.parse(); err != nil {
		return "", err
	}
	dst := bytes.NewBuffer(nil)
	if err := html.Render(dst, p.root); err != nil {
		return "", err
	}
	return dst.String(), nil
}

// Parse parses the page into a root node.
//
// If the root node is already parsed, or has
// errored, the method is a no-op.
func (p *Page) parse() error {
	p.once.Do(func() {
		if p.root, p.err = html.Parse(p.body); p.err != nil {
			p.err = fmt.Errorf("ant: parse html %q - %w", p.URL, p.err)
		}
		p.close()
	})
	return p.err
}

// Query returns all nodes matching selector.
//
// The method returns an empty list if no nodes were found.
func (p *Page) Query(selector string) List {
	var ret List

	if err := p.parse(); err != nil {
		return ret
	}

	if s, err := selectors.Compile(selector); err == nil {
		ret = s.MatchAll(p.root)
	}

	return ret
}

// Text returns the text of the selected node.
//
// The method returns an empty string if the node is not found.
func (p *Page) Text(selector string) string {
	return p.Query(selector).Text()
}

// URLs returns all URLs on the page.
//
// The method skips any invalid URLs.
func (p *Page) URLs() URLs {
	return p.resolve(`a[href]`)
}

// Next all URLs matching the given selector.
func (p *Page) Next(selector string) (URLs, error) {
	return p.resolve(selector), nil
}

// Scan scans data into the given value dst.
func (p *Page) Scan(dst any) error {
	if err := p.parse(); err != nil {
		return err
	}
	return scanner.Scan(dst, p.root, scan.Options{})
}

// Resolve returns resolved URLs matching selector
func (p *Page) resolve(selector string) URLs {
	var anchors = p.Query(selector)
	var ret = make(URLs, 0, len(anchors))

	for _, a := range anchors {
		if href, ok := scan.Attr(a, "href"); ok {
			u, err := url.Parse(href)
			if err != nil {
				continue
			}

			if !u.IsAbs() {
				u = p.URL.ResolveReference(u)
			}

			switch u.Scheme {
			case "https", "http":
				ret = append(ret, u)
			}
		}
	}

	return ret
}

// Close closes the page's body.
func (p *Page) close() error {
	io.Copy(io.Discard, p.body)
	return p.body.Close()
}
