package service

import (
	"reflect"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{
			name: "membersihkan spasi, huruf kecil, dan hapus duplikat",
			tags: []string{"Tugas-Banyak", " dosen-jelas ", "tugas-banyak", "SERU"},
			want: []string{"tugas-banyak", "dosen-jelas", "seru"},
		},
		{
			name: "mengabaikan tag kosong",
			tags: []string{"", "  ", "backend"},
			want: []string{"backend"},
		},
		{
			name: "slice kosong",
			tags: []string{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTags(tt.tags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NormalizeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
