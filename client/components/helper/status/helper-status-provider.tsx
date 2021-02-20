import { APIRequestResult } from '@/api/common';
import { IAPISession, retrieveSessionAgeAPI } from '@/api/session';
import React, { useEffect, useRef, useState } from 'react';
import { HelperStatus, HelperStatusContext } from '../contexts';

export const noExpirationAgeValue = -1;
export const expiredAgeValue = 0;

const useAgeRetrieval = (maxAge: number, initial: number, uuid: string) => {
    const [age, setAge] = useState(initial);
    const ageDecrementTimeout = useRef<NodeJS.Timeout | null>();
    const realAgeRetrievalTimeout = useRef<NodeJS.Timeout | null>();

    useEffect(() => {
        const ageRetrieval = async () => {
            const age = await retrieveSessionAgeAPI(uuid);
            if (age.result === APIRequestResult.FAILED) {
                setAge(() => 0);
            } else {
                setAge(() => age.payload);
                realAgeRetrievalTimeout.current = setTimeout(() => ageRetrieval(), 10000);
            }
        };

        if (maxAge > noExpirationAgeValue && maxAge > expiredAgeValue) {
            ageRetrieval();

            ageDecrementTimeout.current = setInterval(() => {
                setAge(age => age > 0 ? age - 1 : age);
            }, 1000);
        } else {
            setAge(1);
        }

        return () => {
            clearTimeout(realAgeRetrievalTimeout.current);
            clearInterval(ageDecrementTimeout.current);
        }
    }, [uuid]);

    useEffect(() => {
        setAge(initial);
    }, [initial]);

    return { age };
}

export const HelperStatusProvider = (props: React.PropsWithChildren<{ uuid: string, maxAge: number, age: number }>) => {
    const [status, setStatus] = useState(HelperStatus.RUNNING);

    const { age } = useAgeRetrieval(props.maxAge, props.age, props.uuid);

    useEffect(() => {
        if (age === 0) {
            setStatus(HelperStatus.EXPIRED);
        }
    }, [age])

    return <HelperStatusContext.Provider value={{ status, age }}>
        {props.children}
    </HelperStatusContext.Provider>
}