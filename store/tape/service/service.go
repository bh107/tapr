package service

import (
	"os"

	"hpt.space/tapr"
	"hpt.space/tapr/config"
	"hpt.space/tapr/flags"
	"hpt.space/tapr/log"
	"hpt.space/tapr/store"
	"hpt.space/tapr/store/tape"
	"hpt.space/tapr/store/tape/changer"
	"hpt.space/tapr/store/tape/drive"
	"hpt.space/tapr/store/tape/inv"
)

func init() {
	store.Register("tape", New)
}

type service struct {
	name string
	data *dataService
}

type dataService struct {
	inv    inv.Inventory
	chgr   changer.Changer
	drives []drive.Drive
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
	var drvs []drive.Drive
	for _, drv := range cfg.Drives.Read {
		drvOpts := make(map[string]interface{})
		for k, v := range drv.Options {
			drvOpts[k] = v
		}

		drv, err := drive.Create(drv.Driver, drvOpts)
		if err != nil {
			log.Fatal(err)
		}

		drv.Setup(invdb, chgr)

		drvs = append(drvs, drv)
	}

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
