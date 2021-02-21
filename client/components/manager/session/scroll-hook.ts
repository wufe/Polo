import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { FixedSizeList } from "react-window";

export const useScroll = (onLogsProportionChanged: (proportions: number) => void, itemsHeight: number, itemsCount: number, ...deps: any[]) => {
    const contentRef = useRef<HTMLDivElement | null>(null);
    const containerRef = useRef<HTMLDivElement | null>(null);
    const listRef = useRef<FixedSizeList | null>(null);
    const [contentHeight, setContentHeight] = useState(100);
    const [scrolling, setScrolling] = useState(false);
    const [scrollTop, setScrollTop] = useState(0);
    const timeoutIdRef = useRef<NodeJS.Timeout>();
    const [scrollProportions, setScrollProportions] = useState(0);
    const [windowHeight, setWindowHeight] = useState(window.innerHeight);

    const getScrollTop = () => itemsHeight * itemsCount;
    
    useLayoutEffect(() => {
        if (containerRef.current) {
            setContentHeight(containerRef.current.clientHeight);
        }
    }, [containerRef])

    useLayoutEffect(() => {
        const onResize = () => {
            setWindowHeight(window.innerHeight);
            if (containerRef.current) {
                setContentHeight(containerRef.current.clientHeight);
            }
        };
        window.addEventListener('resize', onResize);
        return () => window.removeEventListener('resize', onResize);
    }, []);

    useEffect(() => {
        const container = contentRef.current;
        if (container && !scrolling) {
            if (listRef.current) {
                listRef.current.scrollToItem (itemsCount);
            }
            const clientHeight = container.clientHeight;
            const scrollHeight = getScrollTop();
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
    }, [contentRef.current, windowHeight, scrolling, itemsCount]);

    useEffect(() => {
        return () => clearTimeout(timeoutIdRef.current);
    }, [])

    const onScroll = () => {
        const newScrollTop = contentRef.current.scrollTop;
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

        setScrollTop(newScrollTop);
    }

    return { contentRef, containerRef, listRef, onScroll, contentHeight };
}