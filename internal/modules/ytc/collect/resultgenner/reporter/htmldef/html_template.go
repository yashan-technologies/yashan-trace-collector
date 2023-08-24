package htmldef

import (
	"fmt"
)

const HTML_TEMPLATE = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YTC Report</title>
    %s
</head>
<body>
    <button class="ytc_button" onclick="toggleToc()">显示/隐藏目录</button>
    <div id="catalogs"></div>
    %s
    <script>
        var headings = document.querySelectorAll("h1, h2, h3")
        var toc = "<ul>"

        for (var i = 0; i < headings.length; i++) {
            var heading = headings[i]
            var text = heading.textContent
            var level = parseInt(heading.tagName.charAt(1))

            heading.setAttribute("id", "anchor" + i)

            var listItem = "<li><a href='#anchor" + i + "'>" + text + "</a></li>"

            if (level === 1) {
                toc += "</ul><h2>" + text + "</h2><ul>"
            } else if (level === 2) {
                toc += listItem
            } else if (level === 3) {
                toc += "<ul>" + listItem + "</ul>"
            }
        }

        toc += "</ul>"

        document.getElementById("catalogs").innerHTML = toc
    </script>
    <script>
        function toggleToc () {
            var toc = document.getElementById("catalogs")
            console.log(toc.style.display)
            if (toc.style.display === "none" || (toc.style.display === "")) {
                toc.style.display = "block"
            } else {
                toc.style.display = "none"
            }
        }
    </script>
</body>
</html>
`

func GenHTML(content, style string) string {
	return fmt.Sprintf(HTML_TEMPLATE, style, content)
}