package pkg

type SliceFlags []string

func (i *SliceFlags) String() string {
	return ""
}
func (i *SliceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
