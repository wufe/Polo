import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { ISession, ISessionLog, SessionLogType, SessionStatus } from '@/state/models/session-model';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import { string } from 'mobx-state-tree/dist/internal';
import React, { useEffect, useRef } from 'react';
import { useHistory, useParams } from 'react-router-dom';

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

export const SessionLogs = observer((props: { session: ISession }) => {

    const logsContainerDiv = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        if (logsContainerDiv.current) {
            logsContainerDiv.current.scrollTop = logsContainerDiv.current.scrollHeight;
        }
    }, [logsContainerDiv])

    if (!props.session.logs) return null;

    return <div
        ref={logsContainerDiv}
        className="m-2 shadow-lg text-nord4 py-5 rounded-md border flex-grow flex flex-col bg-nord-5 dark:border-black min-w-0 overflow-x-auto">
        {values(props.session.logs as any as ISessionLog[]).map((log: ISessionLog, key) => {
            const color = colorsByLogType[log.type];
            return <p className="mx-10 leading-relaxed text-sm whitespace-nowrap max-w-full min-w-0 flex items-center" key={key}>
                <span className="uppercase text-xs w-16 min-w-16" style={{ color }}>{log.type}:</span> <span>{log.message}</span>
            </p>
        })}
    </div>
})

type TProps = {
    app: IApp;
}

export const Session = observer((props: TProps) => {

    const { uuid } = useParams<{ uuid: string }>();
    const interval = useRef<NodeJS.Timeout | null>(null);

    const { session } = props.app;

    const history = useHistory();

    useEffect(() => {

        const sessionRetrieval = () => {
            props.app.retrieveSession(uuid)
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
                        alert(request.reason);
                        history.push(`/_polo_/`);
                    } else {
                        interval.current = setTimeout(() => sessionRetrieval(), 1000);
                    }
                });
        };

        sessionRetrieval();
        
        return () => {
            if (interval.current)
                clearTimeout(interval.current);
        }
    }, [uuid])

    useEffect(() => {

        if (session) {
            if (session.status === SessionStatus.STARTED) {
                session.track()
                    .then(request => {
                        if (request.result === APIRequestResult.SUCCEEDED) {
                            location.href = '/';
                        } else {
                            alert(request.reason);
                        }
                    });
            }
        }

    }, [session && session.status]);

    if (!session) return null;

    return <div className="session__component w-10/12 max-w-full flex flex-col mt-3 min-w-0">
        <h1 className="text-4xl mb-3 text-nord1 dark:text-nord5">Session</h1>
        <div className="text-lg text-gray-500 mb-7">Id: {uuid}</div>
        <SessionLogs session={session} />
    </div>
});

export default Session;