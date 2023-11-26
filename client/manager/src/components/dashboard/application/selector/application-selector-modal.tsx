import React from 'react';
import { DefaultModal } from '@/components/modal/default-modal';
import { observer } from 'mobx-react-lite';
import { IApplication } from '@polo/common/state/models';
import {
    DefaultModalDivider,
    DefaultModalHeader,
    DefaultModalItem,
    DefaultModalLayout,
    DefaultModalList,
} from '@/components/modal/default-modal-layout/default-modal-layout';
import {CollectionIcon} from "@heroicons/react/outline";

type TProps = {
    modalName: string;
    applications: IApplication[];

    onApplicationClick: (name: string, index: number) => void;
};
export const ApplicationSelectorModal = observer((props: TProps) => {

    return <DefaultModal name={props.modalName}>
        <DefaultModalLayout>

            <DefaultModalHeader>Applications</DefaultModalHeader>

            <DefaultModalList>
                {props.applications.map((application, index) => <React.Fragment key={application.configuration.hash}>

                    <DefaultModalItem onClick={() => props.onApplicationClick(application.configuration.name, index)}>
                        <CollectionIcon />
                        <span>{application.configuration.name}</span>
                    </DefaultModalItem>

                    <DefaultModalDivider />

                </React.Fragment>)}
            </DefaultModalList>

        </DefaultModalLayout>
    </DefaultModal>
})