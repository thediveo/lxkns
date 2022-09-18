/*
Package all imports and activates all lxkns (container) decorator plugins,
activating them during discoveries.
*/
package all

import (
	_ "github.com/thediveo/lxkns/decorator/composer"                  // pull in decorator plugin
	_ "github.com/thediveo/lxkns/decorator/industrialedge"            // pull in decorator plugin
	_ "github.com/thediveo/lxkns/decorator/kuhbernetes/cricontainerd" // pull in decorator plugin
	_ "github.com/thediveo/lxkns/decorator/kuhbernetes/dockershim"    // pull in decorator plugin
)
