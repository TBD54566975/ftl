package configuration

import (
	"github.com/alecthomas/types/optional"
	"reflect"
	"testing"
)

func TestDecodeSecretRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    *secretRef
		wantErr bool
	}{
		{
			name: "simple with field",
			ref:  "op://development/Access Keys/access_key_id",
			want: &secretRef{
				Vault: "development",
				Item:  "Access Keys",
				Field: optional.Some("access_key_id"),
			},
		},
		{
			name: "simple without field",
			ref:  "op://vault/item",
			want: &secretRef{
				Vault: "vault",
				Item:  "item",
				Field: optional.None[string](),
			},
		},
		{
			name: "lots of spaces",
			ref:  "op://My Awesome Vault/My Awesome Item/My Awesome Field",
			want: &secretRef{
				Vault: "My Awesome Vault",
				Item:  "My Awesome Item",
				Field: optional.Some("My Awesome Field"),
			},
		},
		{
			name:    "missing op://",
			ref:     "development/Access Keys/access_key_id",
			wantErr: true,
		},
		{
			name:    "empty parts",
			ref:     "op://development//access_key_id",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			ref:     "op://development/aws/acce$s",
			wantErr: true,
		},
		{
			name:    "too many parts",
			ref:     "op://development/Access Keys/access_key_id/extra",
			wantErr: true,
		},
		{
			name:    "too few parts",
			ref:     "op://development",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeSecretRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeSecretRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeSecretRef() got = %v, want %v", got, tt.want)
			}
		})
	}
}
