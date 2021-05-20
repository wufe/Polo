import React, { useState } from 'react';
import dayjs from 'dayjs';
import { FailureStatus, TFailuresDictionary } from '@/state/models/failures-model';
import { useFailingSessionsMenuItemDisplay } from './use-failing-sessions-menu-item-display';
import { DefaultModalItem, DefaultModalList, DefaultModalRow } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { ClockIcon } from '@/components/shared/ui-elements/icons/clock/clock-icon';
import { ExclamationCircleIcon } from '@/components/shared/ui-elements/icons/exclamation-circle/exclamation-circle-icon';
import { LeftArrowIcon } from '@/components/shared/ui-elements/icons/left-arrow/left-arrow-icon';
import { TextDocumentIcon } from '@/components/shared/ui-elements/icons/text-document/text-document-icon';
import { ISession } from '@/state/models/session-model';

type TProps = {
    failures: TFailuresDictionary | null;
    onFailingSessionClick: (session: ISession) => void;
}
export const ApplicationOptionsModalFailuresItem = (props: TProps) => {
    const [viewFailingSessions, setViewFailingSession] = useState(false);
    const { anyFailures, anyUnacknowledgedFailures, failuresWithStatus } = useFailingSessionsMenuItemDisplay(props.failures);

    return <>
        {!viewFailingSessions && <DefaultModalItem
            dangerIcon={anyUnacknowledgedFailures} disabled={!anyFailures}
            onClick={() => anyFailures && setViewFailingSession(true)}>
            <ExclamationCircleIcon />
            <span>View failing sessions</span>
        </DefaultModalItem>}

        {viewFailingSessions && <>
            <DefaultModalItem action onClick={() => setViewFailingSession(false)}>
                <LeftArrowIcon />
                <span className="font-bold">Go back</span>
            </DefaultModalItem>

            {failuresWithStatus.map(({ session, status }, index) =>
                <DefaultModalItem
                    key={index}
                    dangerIcon={status === FailureStatus.UNACK}
                    multipleRows
                    onClick={() => props.onFailingSessionClick(session)}>
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
        </>}
    </>
}