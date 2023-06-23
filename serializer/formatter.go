package serializer

type (
	Formatter struct {
		Flatten            bool
		NormalizeKeys      bool
		NameKeyRename      string
		TimestampUnits     string
		TimestampAsRFC3339 bool
	}
)
