import React from 'react';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
dayjs.extend(relativeTime);
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { observer } from 'mobx-react-lite';
import { ISession } from '@/state/models';
import { useHistory } from 'react-router';
import { DefaultModalHeader, DefaultModalItem, DefaultModalLayout, DefaultModalList } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { TFailuresDictionary } from '@/state/models/failures-model';
import { ApplicationOptionsModalFailuresItem } from './failures/application-options-modal-failures-item';
import PencilAltIcon from '@heroicons/react/outline/PencilAltIcon';

type TProps = {
    modalName: string;
    applicationName: string;
    failures: TFailuresDictionary | null;
    onFailingSessionClick: (session: ISession) => void;
    onApplicationConfigurationEditClick: () => void;
};
export const ApplicationOptionsModal = observer((props: TProps) => {

    const history = useHistory();

    return <DefaultModal name={props.modalName}>
        <DefaultModalLayout>
            <DefaultModalHeader>{props.applicationName}</DefaultModalHeader>

            <DefaultModalList>

                <DefaultModalItem onClick={props.onApplicationConfigurationEditClick}>
                    <PencilAltIcon />
                    <span>Edit configuration</span>
                </DefaultModalItem>

                <ApplicationOptionsModalFailuresItem
                    failures={props.failures}
                    onFailingSessionClick={props.onFailingSessionClick} />
            </DefaultModalList>
            

        </DefaultModalLayout>
    </DefaultModal>
})

