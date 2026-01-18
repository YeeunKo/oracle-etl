// Package jsonl은 JSON Lines (NDJSON) 인코딩 기능을 제공합니다.
package jsonl

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_Encode_Map(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name: "간단한 맵 인코딩",
			input: map[string]interface{}{
				"MANDT": "800",
				"VBELN": "0090000001",
				"POSNR": "000010",
			},
			expected: `{"MANDT":"800","POSNR":"000010","VBELN":"0090000001"}`,
		},
		{
			name: "숫자 값 포함",
			input: map[string]interface{}{
				"id":    123,
				"price": 99.99,
			},
			expected: `{"id":123,"price":99.99}`,
		},
		{
			name: "NULL 값 처리",
			input: map[string]interface{}{
				"name":  "test",
				"value": nil,
			},
			expected: `{"name":"test","value":null}`,
		},
		{
			name:     "빈 맵",
			input:    map[string]interface{}{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)

			err := encoder.Encode(tt.input)
			require.NoError(t, err)

			// 버퍼 플러시
			err = encoder.Flush()
			require.NoError(t, err)

			// JSON Lines는 각 줄이 개행문자로 끝남
			result := strings.TrimSuffix(buf.String(), "\n")

			// JSON 키 순서 무관하게 비교
			var expected, actual map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(tt.expected), &expected))
			require.NoError(t, json.Unmarshal([]byte(result), &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestEncoder_Encode_Struct(t *testing.T) {
	type SampleRow struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"created_at"`
	}

	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	testTime := time.Date(2026, 1, 18, 10, 30, 0, 0, time.UTC)
	row := SampleRow{
		ID:        1,
		Name:      "테스트 데이터",
		Active:    true,
		CreatedAt: testTime,
	}

	err := encoder.Encode(row)
	require.NoError(t, err)

	err = encoder.Flush()
	require.NoError(t, err)

	result := strings.TrimSuffix(buf.String(), "\n")

	var decoded SampleRow
	require.NoError(t, json.Unmarshal([]byte(result), &decoded))

	assert.Equal(t, row.ID, decoded.ID)
	assert.Equal(t, row.Name, decoded.Name)
	assert.Equal(t, row.Active, decoded.Active)
	assert.True(t, row.CreatedAt.Equal(decoded.CreatedAt))
}

func TestEncoder_MultipleEncodes(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	rows := []map[string]interface{}{
		{"MANDT": "800", "VBELN": "0090000001", "POSNR": "000010"},
		{"MANDT": "800", "VBELN": "0090000001", "POSNR": "000020"},
		{"MANDT": "800", "VBELN": "0090000002", "POSNR": "000010"},
	}

	for _, row := range rows {
		err := encoder.Encode(row)
		require.NoError(t, err)
	}

	err := encoder.Flush()
	require.NoError(t, err)

	// 결과는 3줄이어야 함
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	assert.Len(t, lines, 3)

	// 각 줄이 유효한 JSON인지 확인
	for i, line := range lines {
		var decoded map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &decoded), "줄 %d 파싱 실패", i)
	}
}

func TestEncoder_Flush(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	err := encoder.Encode(map[string]interface{}{"test": "data"})
	require.NoError(t, err)

	// Flush 호출해도 에러 없어야 함
	err = encoder.Flush()
	require.NoError(t, err)

	// Flush 후 데이터가 버퍼에 있어야 함
	assert.Greater(t, buf.Len(), 0)
}

func TestEncoder_SpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	// 특수 문자, 유니코드, 이스케이프 필요 문자 포함
	input := map[string]interface{}{
		"korean":    "한글 테스트",
		"japanese":  "日本語テスト",
		"newline":   "line1\nline2",
		"tab":       "col1\tcol2",
		"quote":     `value with "quotes"`,
		"backslash": `path\to\file`,
		"emoji":     "테스트 🚀",
	}

	err := encoder.Encode(input)
	require.NoError(t, err)

	err = encoder.Flush()
	require.NoError(t, err)

	result := strings.TrimSuffix(buf.String(), "\n")

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &decoded))

	assert.Equal(t, input["korean"], decoded["korean"])
	assert.Equal(t, input["japanese"], decoded["japanese"])
	assert.Equal(t, input["newline"], decoded["newline"])
	assert.Equal(t, input["tab"], decoded["tab"])
	assert.Equal(t, input["quote"], decoded["quote"])
	assert.Equal(t, input["backslash"], decoded["backslash"])
	assert.Equal(t, input["emoji"], decoded["emoji"])
}

func TestEncoder_LargeData(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	// 274개 컬럼 시뮬레이션 (VBRP 테이블 기준)
	largeRow := make(map[string]interface{})
	for i := 0; i < 274; i++ {
		largeRow[strings.Repeat("COL", 1)+string(rune('A'+i%26))+string(rune('0'+i/26))] = strings.Repeat("X", 100)
	}

	err := encoder.Encode(largeRow)
	require.NoError(t, err)

	err = encoder.Flush()
	require.NoError(t, err)

	result := strings.TrimSuffix(buf.String(), "\n")

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(result), &decoded))

	assert.Equal(t, len(largeRow), len(decoded))
}

func TestEncoder_BytesWritten(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	row := map[string]interface{}{"test": "data"}
	err := encoder.Encode(row)
	require.NoError(t, err)

	bytesWritten := encoder.BytesWritten()
	assert.Greater(t, bytesWritten, int64(0))
	assert.Equal(t, int64(buf.Len()), bytesWritten)
}

func TestEncoder_RowsEncoded(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	for i := 0; i < 5; i++ {
		err := encoder.Encode(map[string]interface{}{"id": i})
		require.NoError(t, err)
	}

	assert.Equal(t, int64(5), encoder.RowsEncoded())
}

func TestStreamEncoder_EncodeRows(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewStreamEncoder(&buf)

	rows := []map[string]interface{}{
		{"id": 1, "name": "first"},
		{"id": 2, "name": "second"},
		{"id": 3, "name": "third"},
	}

	err := encoder.EncodeRows(rows)
	require.NoError(t, err)

	err = encoder.Flush()
	require.NoError(t, err)

	assert.Equal(t, int64(3), encoder.RowsEncoded())

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	assert.Len(t, lines, 3)
}

func TestStreamEncoder_Reset(t *testing.T) {
	var buf bytes.Buffer
	encoder := NewStreamEncoder(&buf)

	err := encoder.Encode(map[string]interface{}{"id": 1})
	require.NoError(t, err)

	err = encoder.Flush()
	require.NoError(t, err)

	assert.Equal(t, int64(1), encoder.RowsEncoded())
	assert.Greater(t, encoder.BytesWritten(), int64(0))

	encoder.Reset()

	assert.Equal(t, int64(0), encoder.RowsEncoded())
	assert.Equal(t, int64(0), encoder.BytesWritten())
}

func BenchmarkEncoder_Encode(b *testing.B) {
	var buf bytes.Buffer
	encoder := NewEncoder(&buf)

	row := map[string]interface{}{
		"MANDT":  "800",
		"VBELN":  "0090000001",
		"POSNR":  "000010",
		"MATNR":  "000000000000123456",
		"ARKTX":  "Test Material Description",
		"KWMENG": 100.5,
		"NETWR":  9999.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encoder.Encode(row)
		buf.Reset()
	}
}
