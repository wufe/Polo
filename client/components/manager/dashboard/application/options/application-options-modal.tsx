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
dayjs.extend(relativeTime);

type TProps = {
    modalName: string;
    applicationName: string;
    failedSessions: ISession[] | null;
    onSessionClick: (session: ISession) => void;
};
export const ApplicationOptionsModal = observer((props: TProps) => {

    const [viewFailingSessions, setViewFailingSession] = useState(false);
    const anyFailedSession = props.failedSessions && props.failedSessions.length > 0;
    const history = useHistory();

    return <DefaultModal name={props.modalName}>
        <DefaultModalLayout>
            <DefaultModalHeader>{props.applicationName}</DefaultModalHeader>

            {!viewFailingSessions && <DefaultModalList>
                <DefaultModalItem
                    dangerIcon={anyFailedSession} disabled={!anyFailedSession}
                    onClick={() => anyFailedSession && setViewFailingSession(true)}>
                    <ExclamationCircleIcon />
                    <span>View failing sessions</span>
                </DefaultModalItem>
            </DefaultModalList>}

            {viewFailingSessions && <DefaultModalList>
                
                <DefaultModalItem action onClick={() => setViewFailingSession(false)}>
                    <LeftArrowIcon />
                    <span className="font-bold">Go back</span>
                </DefaultModalItem>

                {props.failedSessions.map((session, index) =>
                    <DefaultModalItem multipleRows onClick={() => props.onSessionClick(session)} key={index}>
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