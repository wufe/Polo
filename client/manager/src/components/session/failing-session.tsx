import { APIRequestResult } from '@polo/common/api/common';
import { IApp } from '@polo/common/state/models/app-model';
import { ISession, ISessionLog, castAPISessionToSessionModel } from '@polo/common/state/models/session-model';
import dayjs from 'dayjs';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useRef, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { CommitMessage } from '../shared/commit-message';
import { SessionLogs } from './session-logs';
import { useSessionRetrieval } from './session-retrieval-hook';
import { SessionTerminalContainer } from './session-terminal-container';
import '@polo/manager/src/components/session/failing-session.scss';

const useAdvancedTerminal = window.configuration.advancedTerminalOutput;

type TProps = {
    app: IApp;
    uuid: string;
}
export const FailingSession = observer((props: TProps) => {

    const [overlayBottom, setOverlayBottom] = useState(100);
    const [session, setSession] = useState<ISession | null>(null);
    const [logs, setLogs] = useState<ISessionLog[]>([]);
    const history = useHistory();

    const setOverlayProportions = (proportions: number) => {
        const percentage = parseInt(`${proportions * 100}`);
        const inversePercentage = 100 - percentage;
        setOverlayBottom(inversePercentage);
    }

    useEffect(() => {
        Promise.all([
            props.app.failures.retrieveFailedSession(props.uuid),
            props.app.failures.retrieveFailedSessionLogs(props.uuid)
        ])
            .then(([sessionResponse, logsResponse]) => {
                if (sessionResponse.result === APIRequestResult.FAILED) {
                    return history.push(`/_polo_`);
                }
                const session = castAPISessionToSessionModel(sessionResponse.payload);
                setSession(session);
                if (logsResponse.result === APIRequestResult.SUCCEEDED)
                    setLogs(logsResponse.payload);
            })
    }, [props.uuid]);

    if (!props.app.failures.currentSession) return null;

    return <div className="
        mx-auto w-full max-w-6xl flex flex-col min-w-0 min-h-0 flex-1 pt-3 font-quicksand" style={{ maxHeight: 'calc(100vh - 120px)' }}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`, zIndex: 1 }}></div>
        <h1 className="text-4xl px-2 lg:px-0 mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">
            Failing session
        </h1>
        <div className="text-lg text-nord1 dark:text-nord5 mb-4 z-10 border-l pl-3 border-gray-500">
            <span>{props.app.failures.currentSession.displayName}</span>
        </div>
        <CommitMessage {...props.app.failures.currentSession} maxHeight />
        {useAdvancedTerminal && <SessionTerminalContainer app={props.app} session={props.app.failures.currentSession} />}
        {!useAdvancedTerminal && <SessionLogs
            failed
            logs={logs}
            onLogsProportionChanged={setOverlayProportions} />}
    </div>
});

export default FailingSession;