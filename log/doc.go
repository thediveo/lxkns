/*

Package log allows consumers of the lxkns module to forward logging originating
in the lxkns module to whatever logger module they prefer.

The accompanying package github.com/thediveo/log/logrus can be used as a simple
means to forward all lxkns-originating logging messages to the standard logrus
logger instance. YMMV, however.

To forward lxkns-originating log messages to a different logging module only
needs an adaptor implementing the Logger interface: it consists of just two
methods:

    Log(level Level, msg string)
    SetLevel(level Level)

A consumer of the lxkns module then sets her specific logging adapter:

    SetLog(myadapter)

An adapter can either forward the SetLevel() method, but it is also perfectly
fine to ignore these interface calls: this allows keeping separate logging
levels for lxkns and the consuming application; for instance, disabling or
restricting lxkns logging while still doing verbose application logging.

*/
package log
