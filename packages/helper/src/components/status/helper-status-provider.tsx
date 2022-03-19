import {ComponentChildren} from 'preact';
import {useEffect, useRef, useState} from 'preact/hooks';
import { APIRequestResult } from '@polo/common/api/common';
import { IAPISession, retrieveSessionStatusAPI } from '@polo/common/api/session';
import { SessionStatus, SessionKillReason } from '@polo/common/state/models/session-model-enums';
import { HelperStatus, HelperStatusContext } from '../contexts';

export const noExpirationAgeValue = -1;
export const expiredAgeValue = 0;

type TInitialSessionStatus = {
    age       : number;
    status    : SessionStatus;
    killReason: SessionKillReason;
    replacedBy: string;
}

const useStatusRetrieval = (uuid: string, initial: TInitialSessionStatus) => {
    const [status, setStatus] = useState(initial);
    const ageDecrementTimeout = useRef<number | null>();
    const realAgeRetrievalTimeout = useRef<number | null>();

    useEffect(() => {
        const ageRetrieval = async () => {
            const status = await retrieveSessionStatusAPI(uuid);
            if (status.result === APIRequestResult.FAILED) {
                setStatus(s => ({ ...s, age: 0 }));
            } else {
                setStatus(() => status.payload);
                realAgeRetrievalTimeout.current = setTimeout(() => ageRetrieval(), 10000);
            }
        };

        if (status.age > noExpirationAgeValue && status.age > expiredAgeValue) {
            ageRetrieval();

            ageDecrementTimeout.current = setInterval(() => {
                setStatus(s => ({
                    ...s,
                    age: s.age > 0 ? s.age - 1 : s.age
                }));
            }, 1000);
        } else {
            // Here status.age is at most min(noExpirationAgeValue, expiredAgeValue).
            //
            // It means that if noExpirationAgeValue is set to -1 and expiredAgeValue is 0,
            // the status.age value is <= -1.
            // 
            // Age is being set to 1 just to avoid triggering "HelperStatus.EXPIRED"
            // in HelperStatusProvider component.
            setStatus(s => ({...s, age: 1 }));
        }

        return () => {
            clearTimeout(realAgeRetrievalTimeout.current);
            clearInterval(ageDecrementTimeout.current);
        }
    }, [uuid]);

    useEffect(() => {
        setStatus(initial);
    }, [initial]);

    useEffect(() => {
        if (status.status === SessionStatus.STOPPED) {
            clearTimeout(realAgeRetrievalTimeout.current);
            clearInterval(ageDecrementTimeout.current);
            setStatus(s => ({
                ...s,
               age: 0 
            }));
        }
    }, [status.status]);

    return status;
}

export const HelperStatusProvider = (props: { uuid: string, initial: TInitialSessionStatus, children: ComponentChildren }) => {
    const [helperStatus, setHelperStatus] = useState(HelperStatus.RUNNING);

    const { age, killReason, replacedBy, status } = useStatusRetrieval(props.uuid, props.initial);

    useEffect(() => {
        if (age === 0) {
            if (killReason === SessionKillReason.REPLACED) {
                setHelperStatus(HelperStatus.REPLACED);
            } else {
                setHelperStatus(HelperStatus.EXPIRED);
            }
        }
    }, [age])

    return <HelperStatusContext.Provider value={{ helperStatus, age, replacedBy, status, killReason }}>
        {props.children}
    </HelperStatusContext.Provider>
}