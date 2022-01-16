package ruler

import (
	"bytes"
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

func replaceHTML(src string) string {
	root, err := html.Parse(strings.NewReader(src))
	if err != nil {
		panic(err)
	}
	replace(root)

	var b bytes.Buffer
	html.Render(&b, root)
	return b.String()
}

func replace(n *html.Node) {
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
			n1 := &html.Node{
				Type: html.TextNode,
				Data: dataID,
			}
			p := n.Parent
			p.RemoveChild(n)
			p.AppendChild(n1)
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		replace(child)
	}
}
