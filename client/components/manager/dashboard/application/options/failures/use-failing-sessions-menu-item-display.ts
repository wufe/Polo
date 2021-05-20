import dayjs from 'dayjs';
import { FailureStatus, TFailuresDictionary } from "@/state/models/failures-model";
import { ISession } from "@/state/models/session-model";

export const useFailingSessionsMenuItemDisplay = (failures: TFailuresDictionary | null) => {
    const anyFailures = failures && (
        failures[FailureStatus.ACK].length > 0 ||
        failures[FailureStatus.UNACK].length > 0
    );
    const anyUnacknowledgedFailures = failures &&
        failures[FailureStatus.UNACK].length > 0;

    const failuresWithStatus = sortSessionsByCreationTimeDesc(failures);

    return { anyFailures, anyUnacknowledgedFailures, failuresWithStatus };
}

type FailureWithStatus = {
    session: ISession;
    status: FailureStatus;
};
function sortSessionsByCreationTimeDesc(failures: TFailuresDictionary): FailureWithStatus[] {
    if (!failures) return [];

    const sessions: FailureWithStatus[] = [];
    for (const session of failures.acknowledged) {
        sessions.push({
            session,
            status: FailureStatus.ACK,
        });
    }
    for (const session of failures.unacknowledged) {
        sessions.push({
            session,
            status: FailureStatus.UNACK,
        });
    }
    return sessions
        .sort((a, b) => {
            const dateA = dayjs(a.session.createdAt);
            const dateB = dayjs(b.session.createdAt);
            if (dateA.isBefore(dateB))
                return 1;
            if (dateA.isAfter(dateB))
                return -1;
            return 0;
        })
}