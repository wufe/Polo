import { APIPayload, APIRequestResult } from "@/api/common";
import { ISession, ISessionLog } from "@/state/models";
import { values } from "mobx";
import { useRef, useEffect } from "react";
import { useHistory } from "react-router-dom";

export const useSessionRetrieval = (
    retrieveFailedSession: (uuid: string) => Promise<APIPayload<ISession>>,
    onSessionFail: () => void,
    session: ISession,
) => {
    const interval = useRef<NodeJS.Timeout | null>(null);
    const history = useHistory();

    useEffect(() => {

        const sessionStatusRetrieval = async () => {

            const logs: ISessionLog[] = values(session.logs) as any;

            let lastLogUUID: string | undefined = undefined;

            if (logs.length) {
                lastLogUUID = logs[logs.length - 1].uuid;
            }

            let fetchFailed = false;
            try {
                const logsRequest = await session.retrieveLogsAndStatus(lastLogUUID);
                if (logsRequest.result === APIRequestResult.SUCCEEDED) {
                    interval.current = setTimeout(() => sessionStatusRetrieval(), 1000);
                } else {
                    fetchFailed = true;
                }
            } catch (e) {
                console.error(e);
                fetchFailed = true;
            }

            let redirectToDashboard = false;
            if (fetchFailed) {
                redirectToDashboard = true;
                try {
                    const failedSessionRequest = await retrieveFailedSession(session.uuid);
                    if (failedSessionRequest.result === APIRequestResult.SUCCEEDED) {
                        redirectToDashboard = false;
                        onSessionFail();
                    }
                } catch (e) {
                    console.error(e);
                }
            }

            if (redirectToDashboard)
                history.push(`/_polo_/`);
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