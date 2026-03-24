package provider

func nilIfUnknown[T any](val interface {
	IsUnknown() bool
}, getter func() *T) *T {
	if val.IsUnknown() {
		return nil
	}
	return getter()
}
