package server

type Store struct {
	data map[string]string
}

func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

func (s *Store) Set(key string, value string) {

	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) Delete(key string) bool {
	_, existed := s.data[key]

	delete(s.data, key)
	return existed

}
