package detailedlogs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
)

// Record is one structured line extracted from a detailed log: a
// method marker or a standalone JSON payload. Everything else is
// noise and never leaves the parser.
type Record struct {
	Kind    string // "method" or "payload"
	Method  string // for method records: e.g. DeckGetDeckSummariesV3
	Inbound bool   // <== is inbound, ==> is outbound
	Payload []byte // for payload records: the raw JSON line
}

var methodLine = regexp.MustCompile(
	`\[UnityCrossThreadLogger\](==>|<==) ([A-Za-z._0-9]+)`)

// ParseRecords extracts the structured records of a log excerpt; the
// input is only ever an accepted fixture or the local client's log.
func ParseRecords(data []byte) []Record {
	var records []Record
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 64*1024), 4<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		if match := methodLine.FindSubmatch(line); match != nil {
			records = append(records, Record{Kind: "method",
				Method: string(match[2]), Inbound: string(match[1]) == "<=="})
			continue
		}
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 1 && trimmed[0] == '{' && json.Valid(trimmed) {
			records = append(records, Record{Kind: "payload",
				Payload: append([]byte(nil), trimmed...)})
		}
	}
	return records
}

var collectionPair = regexp.MustCompile(`"(\d{5,6})":(\d{1,2})`)

// ExtractCollection pulls the grpId->count map out of a log excerpt
// when a client version that logs it (see Probe) produced one; the
// current client does not, and then ok is false.
func ExtractCollection(data []byte) (map[int]int, bool) {
	loc := collectionShape.FindIndex(data)
	if loc == nil {
		return nil, false
	}
	// The shape match requires trailing commas, so the final pair sits
	// just past it — extend the region to the closing brace.
	region := data[loc[0]:loc[1]]
	if end := bytes.IndexByte(data[loc[1]:], '}'); end >= 0 {
		region = data[loc[0] : loc[1]+end]
	}
	cards := make(map[int]int)
	for _, pair := range collectionPair.FindAllSubmatch(region, -1) {
		grp, errGrp := strconv.Atoi(string(pair[1]))
		count, errCount := strconv.Atoi(string(pair[2]))
		if errGrp == nil && errCount == nil {
			cards[grp] = count
		}
	}
	return cards, len(cards) > 0
}
