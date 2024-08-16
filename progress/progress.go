package progress

type BytesRead func(read uint64) error
type BytesReadWithTotal func(read, total uint64) error
