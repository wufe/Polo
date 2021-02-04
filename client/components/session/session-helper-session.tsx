import { APIRequestResult } from '@/api/common';
import { ISession, IStore } from '@/state/models';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useRef, useState } from 'react';
import './session-helper-session.scss';

export const SessionHelperSessionUUID = observer((props: { uuid: string }) => {
    return <>
        Session: {props.uuid}
    </>
})

type TProps = {
    session: ISession;
}

export const SessionHelperSession = observer((props: TProps) => {
    const realAgeRetrievalTimeout = useRef<NodeJS.Timeout | null>(null);
    const ageDecrementTimeout = useRef<NodeJS.Timeout | null>(null);
    const [age, setAge] = useState(props.session.maxAge);

    const detach = async () => {
        const untrack = await props.session.untrack();
        if (untrack.result === APIRequestResult.SUCCEEDED) {
            location.href = '/';
        }
    }

    useEffect(() => {
        const sessionAgeRetrieval = () => {
            props.session.retrieveAge()
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
                        alert('Session expired');
                        if (ageDecrementTimeout.current)
                            clearInterval(ageDecrementTimeout.current);
                        location.href = '/';
                    } else {
                        realAgeRetrievalTimeout.current = setTimeout(() => sessionAgeRetrieval(), 3000);
                    }
                });
        };
        sessionAgeRetrieval();

        ageDecrementTimeout.current = setInterval(() => {
            setAge(age => age > 0 ? age - 1 : age);
        }, 1000);

        return () => {
            if (realAgeRetrievalTimeout.current)
                clearTimeout(realAgeRetrievalTimeout.current);
            if (ageDecrementTimeout.current)
                clearInterval(ageDecrementTimeout.current);
        }
    }, [props.session.uuid]);

    useEffect(() => {
        setAge(props.session.maxAge);
    }, [props.session.maxAge])

    return <div className="session-helper-session__component">
        <SessionHelperSessionUUID uuid={props.session.uuid} />
        <br />
        Expires in: {age}s
        <br />
        <span className="__detach" onClick={detach}>Detach</span>
    </div>
})