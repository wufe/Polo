import { IApp } from '@/state/models';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useRef } from 'react';
import { useParams } from 'react-router-dom';

type TProps = {
    app: IApp;
}

export const Session = observer((props: TProps) => {

    const { uuid } = useParams<{ uuid: string }>();
    const interval = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        interval.current = setTimeout(() => {
            props.app.retrieveSession(uuid)
        }, 1000)
        return () => {
            if (interval.current)
                clearInterval(interval.current);
        }
    }, [])

    return <div className="session__component">
        <h1>Session</h1>
        <div>Id: {uuid}</div>
    </div>
});

export default Session;