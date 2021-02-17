import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { ISession, ISessionLog, SessionLogType, SessionStatus } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import { string } from 'mobx-state-tree/dist/internal';
import React, { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useHistory, useParams } from 'react-router-dom';
import dayjs from 'dayjs';

const colorsByLogType: {
    [key in SessionLogType]: string;
} = {
    [SessionLogType.TRACE]: "#d48ead",
    [SessionLogType.DEBUG]: "#88c0d0",
    [SessionLogType.INFO]: "#5e81ac",
    [SessionLogType.WARN]: "#ebcb8b",
    [SessionLogType.ERROR]: "#bf616a",
    [SessionLogType.CRITICAL]: "#ad1c2b",
    [SessionLogType.STDIN]: "#AAA",
    [SessionLogType.STDOUT]: "#a3be8c",
    [SessionLogType.STDERR]: "#d08770"
}

const useScroll = (onLogsProportionChanged: (proportions: number) => void, deps: any[]) => {
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
    }, [containerRef.current, windowHeight, ...deps]);

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

export const SessionLogs = observer((props: { logs: ISessionLog[], onLogsProportionChanged: (proportions: number) => void }) => {

    const { containerRef, onScroll } = useScroll(props.onLogsProportionChanged, [props.logs.length]);

    if (!props.logs) return null;

    return <div
        ref={containerRef}
        className="m-2 text-nord-3 dark:text-nord4 py-5 rounded-md flex-grow mt-12 mb-36 flex flex-col min-w-0 min-h-0 overflow-x-hidden no-horizontal-scrollbar"
        style={{ scrollBehavior: 'smooth' }}
        onScroll={onScroll}>
        {props.logs.map((log: ISessionLog, key) => {
            const color = colorsByLogType[log.type];
            return <p className="mx-10 leading-relaxed text-sm whitespace-nowrap max-w-full min-w-0 flex items-center" key={key}>
                <span className="uppercase text-xs font-mono min-w-24 px-3">[{dayjs(log.when).format('HH:mm:ss')}]</span><span className="uppercase text-xs w-16 min-w-16" style={{ color }}>{log.type}:</span> <span>{log.message}</span>
            </p>
        })}
    </div>
})

type TProps = {
    session: ISession;
}
export const Session = observer((props: TProps) => {

    const interval = useRef<NodeJS.Timeout | null>(null);
    const history = useHistory();
    const [overlayBottom, setOverlayBottom] = useState(100);

    useEffect(() => {

        const sessionStatusRetrieval = () => {

            const logs: ISessionLog[] = values(props.session.logs) as any;

            let lastLogUUID: string | undefined = undefined;

            if (logs.length) {
                lastLogUUID = logs[logs.length - 1].uuid;
            }

            props.session.retrieveLogsAndStatus(lastLogUUID)
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
                        history.push(`/_polo_/`);
                    } else {
                        interval.current = setTimeout(() => sessionStatusRetrieval(), 1000);
                    }
                });
        };

        sessionStatusRetrieval();

        return () => {
            if (interval.current)
                clearTimeout(interval.current);
        }
    }, [])

    const setOverlayProportions = (proportions: number) => {
        const percentage = parseInt(`${proportions * 100}`);
        const inversePercentage = 100 - percentage;
        setOverlayBottom(inversePercentage);
    }

    return <div className="
        mx-auto w-10/12 max-w-full flex flex-col min-w-0 min-h-0 flex-1 pt-3" style={{height:'calc(100vh - 120px)'}}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`}}></div>
        <h1 className="text-4xl mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">Session</h1>
        <div className="text-lg text-gray-500 mb-7 font-quicksand z-10">Id: {props.session.uuid}</div>
        <SessionLogs
            logs={values(props.session.logs) as any as ISessionLog[]}
            onLogsProportionChanged={setOverlayProportions} />
    </div>
});

export default Session;