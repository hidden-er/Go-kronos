package txs

import (
	"reflect"
	"testing"
)

func TestExtractTransactionDetails(t *testing.T) {
	tests := []struct {
		name      string
		tx        string
		want      *Transaction
		wantError bool
	}{
		{
			name: "Valid Transaction",
			tx:   "<Dummy TX: PT2PCFTPBGTGT3V2TYR8BCXVJPY7P5UY, Userset: 10, Input Shard: [1], Input Valid: [0], Output Shard: 2, Output Valid: 0 >",
			want: &Transaction{
				InputShard:  []int{1},
				InputValid:  []int{0},
				OutputShard: 2,
				OutputValid: 0,
			},
			wantError: false,
		},
		{
			name:      "Invalid Transaction",
			tx:        "<Invalid TX: Missing Fields>",
			want:      nil,
			wantError: true,
		},
		{
			name: "Valid Transaction with Multiple Inputs",
			tx:   "<Dummy TX: PT2PCFTPBGTGT3V2TYR8BCXVJPY7P5UY, Userset: 10, Input Shard: [1 2 3], Input Valid: [0 1 0], Output Shard: 5, Output Valid: 1 >",
			want: &Transaction{
				InputShard:  []int{1, 2, 3},
				InputValid:  []int{0, 1, 0},
				OutputShard: 5,
				OutputValid: 1,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTransactionDetails(tt.tx)
			if (err != nil) != tt.wantError {
				t.Errorf("extractTransactionDetails() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTransactionDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIntList(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want []int
	}{
		{
			name: "Single Number",
			str:  "1",
			want: []int{1},
		},
		{
			name: "Multiple Numbers",
			str:  "1, 2, 3",
			want: []int{1, 2, 3},
		},
		{
			name: "Spaces and Commas",
			str:  " 1 ,  2,3 ",
			want: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIntList(tt.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseIntList() = %v, want %v", got, tt.want)
			}
		})
	}
}
