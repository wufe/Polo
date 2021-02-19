import { APIRequestResult } from '@/api/common';
import { retrieveSessionAgeAPI } from '@/api/session';
import React, { useEffect, useRef, useState } from 'react';
import { HelperStatus, HelperStatusContext } from '../contexts';

const useAgeRetrieval = (initial: number, uuid: string) => {
    const [age, setAge] = useState(initial);
    const ageDecrementTimeout = useRef<NodeJS.Timeout | null>();
    const realAgeRetrievalTimeout = useRef<NodeJS.Timeout | null>();

    useEffect(() => {
        const ageRetrieval = async () => {
            const age = await retrieveSessionAgeAPI(uuid);
            if (age.result === APIRequestResult.FAILED) {
                setAge(() => -1);
            } else {
                setAge(() => age.payload);
                realAgeRetrievalTimeout.current = setTimeout(() => ageRetrieval(), 10000);
            }
        };
        ageRetrieval();

        ageDecrementTimeout.current = setInterval(() => {
            setAge(age => age > 0 ? age - 1 : age);
        }, 1000);

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

export const HelperStatusProvider = (props: React.PropsWithChildren<{ uuid: string, age: number }>) => {
    const [status, setStatus] = useState(HelperStatus.RUNNING);

    const { age } = useAgeRetrieval(props.age, props.uuid);

    useEffect(() => {
        if (age <= 0) {
            setStatus(HelperStatus.EXPIRED);
        }
    }, [age])

    return <HelperStatusContext.Provider value={{ status, age }}>
        {props.children}
    </HelperStatusContext.Provider>
}