package htmldef

import (
	"encoding/json"
	"fmt"

	"ytc/utils/stringutil"
)

const _graph_template = `
<script>
    Morris.Line({
        element: '%s',
        data: %s,
        xkey: '%s',
        ykeys: %s,
        labels: %s,
        xLabelAngle: 60,
        smooth: true,
        dataLabels: false,
        lineColors: ['#f5b31a', '#4a9f49', '#f78427', '##ed5151', '#1677ff', '#222529'],
    });
</script>`

func GenGraphData(uniqueName string, data []map[string]interface{}, xKey string, yKeys []string, yLabels []string) string {
	dataJSON, _ := json.Marshal(data)
	yKeysJSON, _ := json.Marshal(yKeys)
	yLabelsJSON, _ := json.Marshal(yLabels)
	return fmt.Sprintf(_graph_template,
		uniqueName,
		dataJSON,
		xKey,
		yKeysJSON,
		yLabelsJSON,
	) + stringutil.STR_NEWLINE
}

func GenGraphElement(uniqueName string) string {
	return fmt.Sprintf(`<div id="%s" class="graph"></div>`, uniqueName) + stringutil.STR_NEWLINE
}
