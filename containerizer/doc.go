/*

Package containerizer provides the implementations to store data about
Containers and ContainerEngines. Subpackages implement Containerizers that
discover Containers from certain container engines.

Nota bene: the lxkns model defines only the Container and ContainerEngine
interfaces instead of implementations in order to break import cycles that
otherwise would invariably occur.

*/
package containerizer
