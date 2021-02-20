import { useEffect, useLayoutEffect, useRef, useState } from "react";

export const useScroll = (onLogsProportionChanged: (proportions: number) => void, deps: any[]) => {
    const containerRef = useRef<HTMLDivElement | null>(null);
    const [scrolling, setScrolling] = useState(false);
    const [scrollTop, setScrollTop] = useState(0);
    const timeoutIdRef = useRef<NodeJS.Timeout>();
    const [scrollProportions, setScrollProportions] = useState(0);
    const [windowHeight, setWindowHeight] = useState(window.innerHeight);

    useLayoutEffect(() => {
        const onResize = () => {
            setWindowHeight(window.innerHeight);
        };
        document.addEventListener('resize', onResize);
        return () => document.removeEventListener('resize', onResize);
    }, []);

    useEffect(() => {
        const container = containerRef.current;
        if (container && !scrolling) {
            container.scrollTop = container.scrollHeight;
            const clientHeight = container.clientHeight;
            const scrollHeight = container.scrollHeight;
            if (scrollHeight > clientHeight) {
                // Full height: 100%
                setScrollProportions(1)
                onLogsProportionChanged(1);
            } else {
                let contentHeight = 0;
                for (const child of container.children) {
                    contentHeight += child.clientHeight;
                }
                const proportion = contentHeight / clientHeight;
                setScrollProportions(proportion);
                onLogsProportionChanged(proportion);
            }
        }
    }, [containerRef.current, windowHeight, scrolling, ...deps]);

    useEffect(() => {
        return () => clearTimeout(timeoutIdRef.current);
    }, [])

    const onScroll = () => {
        const newScrollTop = containerRef.current.scrollTop;
        if (scrollTop > newScrollTop || scrolling) {
            setScrolling(true);
            onLogsProportionChanged(0);
            clearTimeout(timeoutIdRef.current)
            timeoutIdRef.current = setTimeout(() => {
                setScrolling(false);
                setScrollProportions(scrollProportions);
                onLogsProportionChanged(scrollProportions);
            }, 5000);
        }

        setScrollTop(containerRef.current.scrollTop);
    }

    return { containerRef, onScroll };
}