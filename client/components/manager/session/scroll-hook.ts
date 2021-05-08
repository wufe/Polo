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
    const [downArrowVisible, setDownArrowVisible] = useState(false);

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
        const clientHeight = contentRef.current.clientHeight;
        const scrollHeight = contentRef.current.scrollHeight;
        // If scrolling up
        if (scrollTop > newScrollTop || scrolling) {
            setScrolling(true);
            clearTimeout(timeoutIdRef.current);
            // Slide the overlay up
            onLogsProportionChanged(0);
        }

        setScrollTop(newScrollTop);

        // How close to the edge the scroll must be
        // to hide the down arrovw
        const thresholdBuffer = 40;

        // Hide the arrow if near the bottom edge
        if (newScrollTop + clientHeight + thresholdBuffer < scrollHeight) {
            setDownArrowVisible(true);
        } else {
            setDownArrowVisible(false);
        }
    }

    // Entering the container, the timer needs to get reset
    const onMouseEnter = () => {
        if (!scrolling) return;
        clearTimeout(timeoutIdRef.current);
    }

    // Moving inside the container, the timer needs to get reset
    const onMouseMove = () => {
        if (!scrolling) return;
        clearTimeout(timeoutIdRef.current);
    }

    // Leaving the div with the mouse means
    // starting the timer which will reset the automatic div scroll
    // (like clicking the down arrow)
    const onMouseLeave = () => {
        if (!scrolling) return;
        clearTimeout(timeoutIdRef.current)
        timeoutIdRef.current = setTimeout(() => {
            setScrolling(false);
            setScrollProportions(scrollProportions);
            onLogsProportionChanged(scrollProportions);
        }, 20000);
    }

    // Clicking the down arrow:
    // - the "scrolling" is set to false
    // - the proportions calculation is refreshed to show the shadow again
    const onDownArrowClick = () => {
        if (!downArrowVisible) return;
        setScrolling(false);
        setScrollProportions(scrollProportions);
        onLogsProportionChanged(scrollProportions);
    }

    return { contentRef, containerRef, listRef, onScroll, onMouseEnter, onMouseMove, onMouseLeave, contentHeight, downArrowVisible, onDownArrowClick };
}