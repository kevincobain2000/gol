package pkg

type SliceFlags []string

// provide a String() method on the type so we can use it with flag.Var
func (i *SliceFlags) String() string {
	return ""
}

// provide a Set() method on the type so we can use it with flag.Var
func (i *SliceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
