package kv

import "github.com/cleung2010/go-git2consul/repository"

// Handles the update of a particular repository by comparing diffs against
// the KV
func (h *KVHandler) HandleUpdate(repo *repository.Repository) error {
	repo.Lock()
	defer repo.Unlock()

	head, err := repo.Head()
	if err != nil {
		return err
	}
	b, err := head.Branch().Name()
	if err != nil {
		return err
	}

	// log.Debugf("(consul) KV GET ref for %s/%s", repo.Name(), b)
	kvRef, err := h.getKVRef(repo, b)
	if err != nil {
		return err
	}

	// Local ref
	localRef := head.Target().String()
	// log.Debugf("(consul) kvRef: %s | localRef: %s", kvRef, localRef)

	if len(kvRef) == 0 {
		// log.Debugf("(consul) KV PUT changes for %s/%s", repo.Name(), b)
		err := h.putBranch(repo, head.Branch())
		if err != nil {
			return err
		}

		err = h.putKVRef(repo, b)
		if err != nil {
			return err
		}
		// log.Debugf("(consul) KV PUT ref for %s/%s", repo.Name(), b)
	} else if kvRef != localRef {
		// Check if the ref belongs to that repo
		err := repo.CheckRef(kvRef)
		if err != nil {
			return err
		}

		// Handle modified and deleted files
		deltas, err := repo.DiffStatus(kvRef)
		if err != nil {
			return err
		}
		h.handleDeltas(repo, deltas)

		err = h.putKVRef(repo, b)
		if err != nil {
			return err
		}
		// log.Debugf("(consul) KV PUT ref for %s/%s", repo.Name(), b)
	}

	return nil
}
