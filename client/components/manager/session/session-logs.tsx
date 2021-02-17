import { ISessionLog, SessionLogType } from '@/state/models';
import dayjs from 'dayjs';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { useScroll } from './scroll-hook';

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