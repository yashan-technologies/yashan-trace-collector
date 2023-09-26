package htmldef

import (
	"fmt"
)

const _html_template = `
<!DOCTYPE html>
<html  lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YTC Report</title>
    <link rel='stylesheet' href='./ytc_report_static/css/morris.css'>
    <script src='./ytc_report_static/js/raphael.min.js'></script>
    <script src='./ytc_report_static/js/morris.js'></script>
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
            if (toc.style.display === "none" || (toc.style.display === "")) {
                toc.style.display = "block"
            } else {
                toc.style.display = "none"
            }
        }
    </script>
    <button id="back-to-top" class="ytc_button">Top</button>

    <script>
        window.addEventListener('scroll', function () {
            var button = document.querySelector('#back-to-top')
            if (window.pageYOffset > 100) {
                button.style.display = 'block'
            } else {
                button.style.display = 'none'
            }
        })

        var button = document.querySelector('#back-to-top')
        button.addEventListener('click', function () {
            window.scrollTo(0, 0)
        });
    </script>
    <button id="back-to-top">Top</button>
    %s
</body>
</html>
`

func GenHTML(content, graph string) string {
	return fmt.Sprintf(_html_template, _html_css, content, graph)
}
