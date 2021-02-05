import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { ISession, ISessionLog, SessionLogType, SessionStatus } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import { string } from 'mobx-state-tree/dist/internal';
import React, { useEffect, useRef } from 'react';
import { useHistory, useParams } from 'react-router-dom';
import dayjs from 'dayjs';

const colorsByLogType: {
    [key in SessionLogType]: string;
} = {
    [SessionLogType.TRACE]: "#d8dee9",
    [SessionLogType.DEBUG]: "#88c0d0",
    [SessionLogType.INFO]: "#5e81ac",
    [SessionLogType.WARN]: "#ebcb8b",
    [SessionLogType.ERROR]: "#bf616a",
    [SessionLogType.CRITICAL]: "#ad1c2b",
    [SessionLogType.STDIN]: "#eceff4",
    [SessionLogType.STDOUT]: "#a3be8c",
    [SessionLogType.STDERR]: "#d08770"
}

export const SessionLogs = observer((props: { logs: ISessionLog[] }) => {

    const logsContainerDiv = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        if (logsContainerDiv.current) {
            logsContainerDiv.current.scrollTop = logsContainerDiv.current.scrollHeight;
        }
    }, [logsContainerDiv.current, props.logs.length])

    if (!props.logs) return null;

    return <div
        ref={logsContainerDiv}
        className="m-2 shadow-lg text-nord4 py-5 rounded-md border flex-grow flex flex-col bg-nord-5 dark:border-black min-w-0 min-h-0 overflow-x-auto">
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
                        alert(request.reason);
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

    return <div className="mx-auto w-10/12 max-w-full flex flex-col pt-3 min-w-0 min-h-0 max-h-screen">
        <h1 className="text-4xl mb-3 text-nord1 dark:text-nord5">Session</h1>
        <div className="text-lg text-gray-500 mb-7">Id: {props.session.uuid}</div>
        <SessionLogs logs={values(props.session.logs) as any as ISessionLog[]} />
    </div>
});

export default Session;