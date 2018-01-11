package config_test

import (
	_ "hpt.space/tapr/store/fs"
	_ "hpt.space/tapr/store/tape"
)

const rawConfig = `
stores: {
  "debug": {
    backend: "fs",
    
    root: "/tmp/tapr"
  },

  "archive": {
    backend: "tape",

    cleaning-prefix: "CLN",

    db: {
      name: "tapr",
      username: "tapr",
      password: "secret"
    },

    changers: {
      "primary": {
        driver: "fake",
        options: {
          transfer: 4,
          storage: 32,
          ix: 4,
          volumes: 16,
        }
      },

      "secondary": {
        driver: "mtx",
        options: {
          path: "/dev/changer",
        }
      }
    },

    drives: {
      "/dev/st0": {
        class: "read"
      },

      "/dev/st1": {
        class: "write"
      },

      "/dev/st2": {
        class: "read"
      },

      "/dev/st3": {
        class: "write"
      }
    }
  }
}
`
