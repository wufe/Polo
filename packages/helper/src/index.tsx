const inIFrame = (() => {
    try {
        return window.self !== window.top;
    } catch (e) {
        return true;
    }
})();

if (!inIFrame) ready(() => import('./components/app'));

function ready(fn: () => void) {
    if (document.readyState !== 'loading') {
        fn();
    } else {
        document.addEventListener('DOMContentLoaded', fn);
    }
}