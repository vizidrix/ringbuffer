package ringbuffer

type BatchStates uint64

const (
	AVAILABLE BatchStates = 1
	WRITING               = 2
	CANCELED              = 3
	PUBLISHED             = 4
)

func (state BatchStates) String() string {
	switch state {
	case AVAILABLE:
		{
			return "AVAILABLE"
		}
	case WRITING:
		{
			return "WRITING"
		}
	case CANCELED:
		{
			return "CANCELED"
		}
	case PUBLISHED:
		{
			return "PUBLISHED"
		}
	}
	return "INVALID"

}
