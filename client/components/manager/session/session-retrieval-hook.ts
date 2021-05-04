import { APIRequestResult } from "@/api/common";
import { ISession, ISessionLog } from "@/state/models";
import { values } from "mobx";
import { useRef, useEffect } from "react";
import { useHistory } from "react-router-dom";

export const useSessionRetrieval = (session: ISession) => {
    const interval = useRef<NodeJS.Timeout | null>(null);
    const history = useHistory();

    useEffect(() => {

        const sessionStatusRetrieval = () => {

            const logs: ISessionLog[] = values(session.logs) as any;

            let lastLogUUID: string | undefined = undefined;

            if (logs.length) {
                lastLogUUID = logs[logs.length - 1].uuid;
            }

            session.retrieveLogsAndStatus(lastLogUUID)
                .then(request => {
                    if (request.result === APIRequestResult.FAILED) {
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
}

export const useFailingSessionRetrieval = (sessino: ISession) => {
    
}