/*

Package all imports and activates all lxkns (container) decorator plugins,
activating them during discoveries.

*/
package all

import (
	_ "github.com/thediveo/lxkns/decorator/composer"
	_ "github.com/thediveo/lxkns/decorator/industrialedge"
	_ "github.com/thediveo/lxkns/decorator/kuhbernetes/cricontainerd"
	_ "github.com/thediveo/lxkns/decorator/kuhbernetes/dockershim"
)
