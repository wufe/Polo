import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { SessionStatus } from '@/state/models/session-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useRef } from 'react';
import { useHistory, useParams } from 'react-router-dom';

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

    return <div className="session__component">
        <h1>Session</h1>
        <div>Id: {uuid}</div>
        <br /><br />
        <div>
            {session && session.logs.map((log, key) => <p key={key}>{log.type}: {log.message}</p>)}
        </div>
    </div>
});

export default Session;