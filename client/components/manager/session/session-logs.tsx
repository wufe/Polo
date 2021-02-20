import { ISessionLog, SessionLogType } from '@/state/models';
import dayjs from 'dayjs';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { useScroll } from './scroll-hook';
const { parse } = require('ansicolor');

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

const parseMessage = (message: string) => {
    try {
        const parsed = parse(message);
        return parsed.spans
            .map(({ css, text }: { css: string; text: string; }, i: number) => {
                const styles = css.split(';').reduce<{ [k: string]: string }>((acc, style) => {
                    if (!style.trim())
                        return acc;
                    const [key, value] = style.trim().split(':');
                    acc[key] = value;
                    return acc;
                }, {});
                return <span style={{...styles, paddingRight: '2px'}} key={i}>{text}</span>
            });
    } catch {
        return <span>{message}</span>
    }
}

export const SessionLogs = observer((props: { logs: ISessionLog[], onLogsProportionChanged: (proportions: number) => void }) => {

    const { containerRef, onScroll } = useScroll(props.onLogsProportionChanged, [props.logs.length]);

    if (!props.logs) return null;

    return <div
        ref={containerRef}
        className="m-2 text-nord-3 dark:text-nord4 py-5 rounded-md flex-grow mt-10 mb-36 flex flex-col min-w-0 min-h-0 overflow-x-hidden no-horizontal-scrollbar"
        style={{ scrollBehavior: 'smooth' }}
        onScroll={onScroll}>
        {props.logs.map((log: ISessionLog, key) => {
            const color = colorsByLogType[log.type];
            return <p className="mx-10 leading-relaxed text-sm whitespace-nowrap max-w-full min-w-0 flex items-center" key={key}>
                <span className="uppercase text-xs font-mono min-w-24 px-3">[{dayjs(log.when).format('HH:mm:ss')}]</span><span className="uppercase text-xs w-16 min-w-16" style={{ color }}>{log.type}:</span>
                {parseMessage(log.message)}
            </p>
        })}
    </div>
})