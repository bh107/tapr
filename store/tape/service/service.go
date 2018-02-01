package service

import (
	"os"
	"sync"

	"tapr.space"
	"tapr.space/config"
	"tapr.space/flags"
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
	data *dataService
}

type dataService struct {
	inv    inv.Inventory
	chgr   changer.Changer
	drives map[string]drive.Drive
}

var _ store.Store = (*service)(nil)

// New creates a new store.Store service.
func New(name string, _cfg config.StoreConfig) (store.Store, error) {
	op := "store/tapr/service.New[" + name + "]"
	cfg := _cfg.Embedded.(tape.Config)

	// setup inventory
	invOpts := cfg.Inventory.Options
	invOpts["cleaning-prefix"] = cfg.CleaningPrefix
	invdb, err := inv.Create(cfg.Inventory.Driver, invOpts)
	if err != nil {
		log.Fatal(err)
	}

	vols, err := invdb.Volumes()
	if err != nil {
		log.Fatal(err)
	}

	if flags.ResetDB {
		log.Debug.Printf("%s: resetting inventory database", op)
		if err := invdb.Reset(); err != nil {
			log.Fatal(err)
		}

		vols = nil
	}

	// setup changer
	chgrOpts := make(map[string]interface{})
	for k, v := range cfg.Changers["primary"].Options {
		chgrOpts[k] = v
	}

	chgrOpts["cleaning-prefix"] = cfg.CleaningPrefix
	chgrOpts["vols"] = vols
	chgr, err := changer.Create(cfg.Changers["primary"].Driver, chgrOpts)
	if err != nil {
		log.Fatal(err)
	}

	if flags.Audit {
		log.Debug.Printf("%s: auditing inventory", op)
		if err := invdb.Audit(chgr); err != nil {
			log.Fatal(err)
		}
	}

	// setup drives
	drvs := map[string]drive.Drive{}
	var wg sync.WaitGroup
	for name, drvCfg := range cfg.Drives.Write {
		drv, err := drive.Create(name, drvCfg.Driver, drvCfg)
		if err != nil {
			log.Fatal(err)
		}

		wg.Add(1)

		go func() {
			drv.Setup(invdb, chgr)
			wg.Done()
		}()

		drvs[name] = drv
	}

	wg.Wait()

	log.Debug.Printf("%s: drives ready", op)

	return &service{
		name: name,
		data: &dataService{
			inv:    invdb,
			chgr:   chgr,
			drives: drvs,
		},
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
	panic("not implemented")
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
