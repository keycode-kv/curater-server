package newsletter

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

func (s NewsLetter) removeElementsFromHTML(ctx context.Context) (parsedHTML string, err error) {

	removeClass := []string{
		"gmail_attr",
		"gmail_signature",
		"gmail_signature_prefix",
	}

	doc, err := html.Parse(strings.NewReader(s.HTMLContent))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	for _, c := range removeClass {
		removeElementsByClass(doc, c)
	}
	var modifiedHTML strings.Builder
	if err = html.Render(&modifiedHTML, doc); err != nil {
		fmt.Println("Error rendering HTML:", err)
		return
	}

	parsedHTML = modifiedHTML.String()
	return
}

func removeElementsByClass(node *html.Node, classToRemove string) {
	if node.Type == html.ElementNode {
		// Check if the element has a "class" attribute
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == classToRemove {
				// Remove the entire node, including its children
				parent := node.Parent
				if parent != nil {
					parent.RemoveChild(node)
				}
				break
			}
		}
	}

	// Recursively process child nodes
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		removeElementsByClass(child, classToRemove)
		child = next
	}
}

func (s NewsLetter) parsePlainText() string {
	substr := strings.Split(s.PlainText, "To:")
	x := substr[len(substr)-1]

	substr2 := strings.SplitN(x, "\n", 2)

	return substr2[1]
}
