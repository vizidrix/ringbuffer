package ringbuffer

import (
	"errors"
	. "launchpad.net/tomb"
	"time"
)

type ClaimResult struct {
	ResultChan chan *Batch
	Tomb       Tomb
}

func (claimResult *ClaimResult) Wait(timeout time.Duration) (*Batch, error) {
	if timeout == 0*time.Nanosecond {
		select { // Don't set a timeout if duration was zero ns
		case result := <-claimResult.ResultChan:
			{
				return result, nil
			}
		case <-claimResult.Tomb.Dying():
			{
				return nil, claimResult.Tomb.Err()
			}
		}
	}
	select {
	case result := <-claimResult.ResultChan:
		{
			return result, nil
		}
	case <-claimResult.Tomb.Dying():
		{
			return nil, claimResult.Tomb.Err()
		}
	case <-time.After(timeout):
		{
			err := errors.New("Claim wait timeout")
			claimResult.Tomb.Kill(err)
			return nil, err
		}
	}
}

// TODO: Need to make sure canceled claims are somehow cleaned up reliably
func (claimResult *ClaimResult) Cancel() {
	// Maybe go routine waiting for result can handle cleanup
	// if it notices the result chan is closed when it tries
	// to publish?
	//claimResult.CloseAll()
	claimResult.Tomb.Kill(errors.New("Claim canceled"))
}
