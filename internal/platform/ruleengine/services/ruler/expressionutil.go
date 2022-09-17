package ruler

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/net/html"
)

//Behaviour is the enum to hold the action behaviour
type Behaviour string

//Types of behaviours
const (
	Create  Behaviour = "create"
	Update  Behaviour = "update"
	Retrive Behaviour = "retrive"
)

//Action is the parsed action expression
type Action struct {
	EntityID  string
	ItemID    string
	SecItemID string
	Behaviour Behaviour
}

//FetchEntityID fetches the root id from the rule
func FetchEntityID(expression string) string {
	parts := strings.Split(expression, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

//FetchItemID fetches item type from the list
func FetchItemID(expression string) string {
	parts := strings.Split(expression, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func ReplaceHTML(src string) string {
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		panic(err)
	}
	ch := &CodeHTML{
		isActive: false,
	}
	replace(root, ch)
	if ch.isActive { // render HTML only for HTML strings
		src = renderNode(root)
	}

	src = strings.Replace(src, `\"`, "", -1)
	src = strings.Replace(src, `<html><head></head><body><p>`, "", -1)
	src = strings.Replace(src, `</p></head></body></html>`, "", -1)
	src = strings.Replace(src, `</p></body></html>`, "", -1)

	return src

}

func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	escaped := html.UnescapeString(buf.String())
	return escaped
}

func replace(n *html.Node, ch *CodeHTML) {
	var dataID string
	if n.Type == html.ElementNode && n.Data == "span" {
		dataID = ""
		for _, a := range n.Attr {
			if a.Key == "data-id" {
				dataID = a.Val
				break
			}
		}
		if dataID != "" {
			ch.isActive = true
			n.Type = html.TextNode
			n.Data = dataID
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		replace(child, ch)
	}
}

type CodeHTML struct {
	isActive bool
}
