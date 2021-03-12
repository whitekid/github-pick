package pocket

type GetOptions struct {
	search   string // Only return items whose title or url contain the search string
	domain   string // Only return items from a particular domain
	favorite int    // only return favorited items
}

type GetOption interface {
	apply(*GetOptions)
}

type funcGetOption struct {
	f func(o *GetOptions)
}

func (f *funcGetOption) apply(o *GetOptions) { f.f(o) }

func newFuncGetOption(f func(o *GetOptions)) GetOption {
	return &funcGetOption{
		f: f,
	}
}

func WithSearch(search string) GetOption {
	return newFuncGetOption(func(o *GetOptions) {
		o.search = search
	})
}

func WithDomain(domain string) GetOption {
	return newFuncGetOption(func(o *GetOptions) {
		o.domain = domain
	})
}
func WithFavorate(favorate int) GetOption {
	return newFuncGetOption(func(o *GetOptions) {
		o.favorite = favorate
	})
}
