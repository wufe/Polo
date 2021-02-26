import { APIRequestResult } from '@/api/common';
import { IApp } from '@/state/models';
import { SessionStatus } from '@/state/models/session-model';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useState } from 'react';
import { useHistory, useParams } from 'react-router-dom';
import { Session } from './session';

type TProps = {
    app: IApp;
}
export const SessionPage = observer((props: TProps) => {

    const params = useParams<{ uuid: string; 0: string }>();
    const [loading, setLoading] = useState(false);
    const catchall = params[0];
    const uuid = params.uuid;
    
    const history = useHistory();
    const { session } = props.app;

    useEffect(() => {
        if (!session || session.uuid !== uuid) {
            props.app.retrieveSession(uuid)
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
                        alert(request.reason);
                        history.push(`/_polo_/`);
                    }
                });
        }
        
    }, [uuid])

    useEffect(() => {
        const onLogsPage = catchall?.startsWith('logs');
        if (session && !onLogsPage) {
            if (session.status === SessionStatus.STARTED) {
                setLoading(true);
                session.track()
                    .then(request => {
                        if (request.result === APIRequestResult.SUCCEEDED) {
                            location.href = `/${catchall}`;
                        }
                    });
            }
        }

    }, [session && session.status]);

    if (!session || session.uuid !== uuid || loading) return null;

    return <Session session={session} />
});

export default SessionPage;