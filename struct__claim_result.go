package ringbuffer

import (
	"errors"
	"log"
	//. "launchpad.net/tomb"
	//"time"
)

type ClaimResult struct {
	ResultChan chan *Batch
	CancelChan chan struct{}
	ErrorChan  chan error
	//Tomb       Tomb
}

func NewClaimResult() *ClaimResult {
	return &ClaimResult{
		ResultChan: make(chan *Batch),
		CancelChan: make(chan struct{}),
		ErrorChan:  make(chan error),
	}
}

func (claimResult *ClaimResult) Wait() (*Batch, error) {
	select { // Don't set a timeout if duration was zero ns
	case result := <-claimResult.ResultChan:
		{
			//log.Printf("Result: %s", result)
			return result, nil
		}
	case <-claimResult.CancelChan:
		{
			log.Printf("Cancel")
			return nil, errors.New("Claim canceled (no timeout)")
		}
	case err := <-claimResult.ErrorChan:
		{
			log.Printf("Err: %s", err)
			return nil, err
		}
	}
}

/*
func (claimResult *ClaimResult) Wait(timeout time.Duration) (*Batch, error) {
	if timeout == 0 {
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
*/
// TODO: Need to make sure canceled claims are somehow cleaned up reliably
func (claimResult *ClaimResult) Cancel() {
	claimResult.CancelChan <- struct{}{}
	// Maybe go routine waiting for result can handle cleanup
	// if it notices the result chan is closed when it tries
	// to publish?
	//claimResult.CloseAll()
	//claimResult.Tomb.Kill(errors.New("Claim canceled"))
}
