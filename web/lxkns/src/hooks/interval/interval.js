import { useEffect, useRef } from 'react'

// useInterval is the declarative cousin to JS' setInternval, using react hooks.
// See: https://overreacted.io/making-setinterval-declarative-with-react-hooks/;
// the posted code is licensed under the MIT license, see:
// https://github.com/gaearon/overreacted.io/blob/master/LICENSE-code-snippets.
const useInterval = (callback, delay) => {
    const savedCallback = useRef() // no useState() here.

    // Whenever the callback or the delay changes, make sure that we remember
    // and use the most recent callback; so, no useState(), but a reference
    // instead.
    useEffect(() => {
        savedCallback.current = callback
    });

    // Whenever the delay changes, we need to set up the interval timer anew.
    // But since the callback might be changed (independently) too, we must use
    // a reference to the callback instead of the "frozen" component state.
    useEffect(() => {
        function tick() {
            savedCallback.current()
        }
        // Only set the interval timer, if the delay value isn't null; any
        // previous timer is automatically removed because we previously told
        // react how to clean it up when changing the delay value. It's like
        // parallel universes...
        if (delay !== null) {
            let id = setInterval(tick, delay)
            // ...and tell react (how) to clean up the old interval timer when
            // the delay value changes.
            return () => clearInterval(id)
        }
    }, [delay])
};

export default useInterval
