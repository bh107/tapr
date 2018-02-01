package service

import (
	"os"
	"sync"

	"tapr.space"
	"tapr.space/config"
	"tapr.space/flags"
	"tapr.space/format"
	"tapr.space/log"
	"tapr.space/store"
	"tapr.space/store/tape"
	"tapr.space/store/tape/changer"
	"tapr.space/store/tape/drive"
	"tapr.space/store/tape/inv"
)

func init() {
	store.Register("store/tape", New)
}

type service struct {
	name string

	inv    inv.Inventory
	chgr   changer.Changer
	drives map[string]*drive.Drive

	fmtr format.Formatter
}

var _ store.Store = (*service)(nil)

// New creates a new store.Store service.
func New(name string, _cfg config.StoreConfig) (store.Store, error) {
	op := "store/tape/service.New[" + name + "]"
	cfg := _cfg.Embedded.(tape.Config)

	// setup inventory
	invOpts := cfg.Inventory.Options
	invOpts["cleaning-prefix"] = cfg.CleaningPrefix
	invdb, err := inv.Create(cfg.Inventory.Driver, invOpts)
	if err != nil {
		log.Fatal(err)
	}

	// reset the database if requested
	if flags.ResetDB {
		log.Debug.Printf("%s: resetting inventory database", op)
		if err := invdb.Reset(); err != nil {
			log.Fatal(err)
		}
	}

	// setup changer
	chgrOpts := make(map[string]interface{})
	for k, v := range cfg.Changers["primary"].Options {
		chgrOpts[k] = v
	}

	chgrOpts["cleaning-prefix"] = cfg.CleaningPrefix

	if flags.EmulateDevices {
		vols, err := invdb.Volumes()
		if err != nil {
			log.Fatal(err)
		}

		chgrOpts["vols"] = vols
	}

	chgr, err := changer.Create(cfg.Changers["primary"].Driver, chgrOpts)
	if err != nil {
		log.Fatal(err)
	}

	// perform an audit if requested
	if flags.Audit {
		log.Debug.Printf("%s: auditing inventory", op)
		if err := invdb.Audit(chgr); err != nil {
			log.Fatal(err)
		}
	}

	fmtr, err := format.Create(cfg.Drives.Format)
	if err != nil {
		log.Fatal(err)
	}

	// setup drives
	var wg sync.WaitGroup
	drvs := make(map[string]*drive.Drive)
	for name, cfg := range cfg.Drives.Write {
		drv, err := drive.New(name, cfg)
		if err != nil {
			log.Fatal(err)
		}

		drvs[name] = drv

		wg.Add(1)

		go func() {
			if err := drv.Start(invdb, chgr, fmtr); err != nil {
				log.Fatal(err)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	log.Debug.Printf("%s: drives ready", op)

	return &service{
		name:   name,
		inv:    invdb,
		chgr:   chgr,
		drives: drvs,
		fmtr:   fmtr,
	}, nil
}

func (s *service) String() string {
	return s.name
}

func (s *service) Create(name tapr.PathName) (tapr.File, error) {
	panic("not implemented")
}

func (s *service) Open(name tapr.PathName) (tapr.File, error) {
	panic("not implemented")
}

func (s *service) OpenFile(name tapr.PathName, flag int) (tapr.File, error) {
	return s.drives["write0"].Storage.OpenFile(name, flag)
}

func (s *service) Append(name tapr.PathName) (tapr.File, error) {
	panic("not implemented")
}

func (s *service) Mkdir(name tapr.PathName) error {
	panic("not implemented")
}

func (s *service) MkdirAll(name tapr.PathName) error {
	panic("not implemented")
}

func (s *service) Stat(name tapr.PathName) (os.FileInfo, error) {
	panic("not implemented")
}
