package promptbuilder

import (
	"fmt"
	"os"
	"strings"
)

// BuildSystemInstructions loads the instruction text only
func BuildSystemInstructions(instructionsPath string) (string, error) {
	instructionsBytes, err := os.ReadFile(instructionsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read instructions: %w", err)
	}
	return string(instructionsBytes), nil
}

// LoadCanonicalData reads all JSON files in the data directory and returns a labeled map
func LoadCanonicalData(dataDir string) (map[string]string, error) {
	data := make(map[string]string)

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, err := os.ReadFile(fmt.Sprintf("%s/%s", dataDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}
		label := strings.TrimSuffix(entry.Name(), ".json")
		data[label] = string(content)
	}

	return data, nil
}

// BuildUserPrompt constructs the full user prompt including instructions, canonical data, writing samples, and the user query
func BuildUserPrompt(query string, chunks []string, canonicalData map[string]string, instructions string) string {
	var canonicalBlocks []string
	for key, value := range canonicalData {
		block := fmt.Sprintf("Here is the JSON data that represents Matt's %s:\n```json\n%s\n```", key, value)
		canonicalBlocks = append(canonicalBlocks, block)
	}

	var labeledChunks []string
	for i, chunk := range chunks {
		labeled := fmt.Sprintf("[Writing sample #%d]\n%s", i+1, chunk)
		labeledChunks = append(labeledChunks, labeled)
	}

	return fmt.Sprintf(`%s

---

%s

---

Here are writing samples to guide tone and style:
%s

---

The user has asked the following question:
%s
`,
		instructions,
		strings.Join(canonicalBlocks, "\n\n"),
		strings.Join(labeledChunks, "\n\n"),
		query,
	)
}
