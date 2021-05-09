import React, { useState } from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { observer } from 'mobx-react-lite';
import { IApplication, ISession } from '@/state/models';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { useHistory } from 'react-router';
import { DefaultModalHeader, DefaultModalItem, DefaultModalLayout, DefaultModalList, DefaultModalRow } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { ExclamationCircleIcon } from '@/components/shared/elements/icons/exclamation-circle/exclamation-circle-icon';
import { LeftArrowIcon } from '@/components/shared/elements/icons/left-arrow/left-arrow-icon';
import { TextDocumentIcon } from '@/components/shared/elements/icons/text-document/text-document-icon';
import { ClockIcon } from '@/components/shared/elements/icons/clock/clock-icon';
import { FailureStatus, TFailuresDictionary } from '@/state/models/failures-model';
dayjs.extend(relativeTime);

type TProps = {
    modalName: string;
    applicationName: string;
    failures: TFailuresDictionary | null;
    onSessionClick: (session: ISession) => void;
};
export const ApplicationOptionsModal = observer((props: TProps) => {

    const [viewFailingSessions, setViewFailingSession] = useState(false);
    const anyFailures = props.failures && (
        props.failures[FailureStatus.ACK].length > 0 ||
        props.failures[FailureStatus.UNACK].length > 0
    );
    const anyUnacknowledgedFailures = props.failures &&
        props.failures[FailureStatus.UNACK].length > 0;
    const history = useHistory();

    const failures = sortSessionsByCreationTimeDesc(props.failures);

    return <DefaultModal name={props.modalName}>
        <DefaultModalLayout>
            <DefaultModalHeader>{props.applicationName}</DefaultModalHeader>

            {!viewFailingSessions && <DefaultModalList>
                <DefaultModalItem
                    dangerIcon={anyUnacknowledgedFailures} disabled={!anyFailures}
                    onClick={() => anyUnacknowledgedFailures && setViewFailingSession(true)}>
                    <ExclamationCircleIcon />
                    <span>View failing sessions</span>
                </DefaultModalItem>
            </DefaultModalList>}

            {viewFailingSessions && <DefaultModalList>
                
                <DefaultModalItem action onClick={() => setViewFailingSession(false)}>
                    <LeftArrowIcon />
                    <span className="font-bold">Go back</span>
                </DefaultModalItem>

                {failures.map(({session, status}, index) =>
                    <DefaultModalItem
                        key={index}
                        dangerIcon={status === FailureStatus.UNACK}
                        multipleRows
                        onClick={() => props.onSessionClick(session)}>
                        <DefaultModalRow>
                            <TextDocumentIcon />
                            <span>{session.commitMessage.split('\n')[0]}</span>
                        </DefaultModalRow>
                        <DefaultModalRow secondary indented>
                            <ClockIcon />
                            <span className="text-gray-400 text-sm">{dayjs(session.createdAt).fromNow()}</span>
                        </DefaultModalRow>
                    </DefaultModalItem>
                )}
            </DefaultModalList>}

        </DefaultModalLayout>
    </DefaultModal>
})

type FailureWithStatus = {
    session: ISession;
    status : FailureStatus;
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