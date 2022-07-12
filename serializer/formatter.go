package serializer

type (
	Formatter struct {
		Flatten            bool   `toml:"flatten"`
		NormalizeKeys      bool   `toml:"normalize_keys"`
		NameKeyRename      string `toml:"name_key_rename"`
		TimestampAsRFC3339 bool   `toml:"timestamp_as_rfc3339"`
	}
)
