import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { DefaultModalDivider, DefaultModalHeader, DefaultModalItem, DefaultModalLayout, DefaultModalList } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { AnnotationIcon } from '@/components/shared/ui-elements/icons/annotation/annotation-icon';
import { CubeIcon } from '@/components/shared/ui-elements/icons/cube/cube-icon';

type TProps = {
    name                       : string;
    checkoutName               : string;
    onSessionCreationSubmission: () => void;
    onCommitMessageSelection      : () => void;
}
export const ApplicationCheckoutModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <DefaultModalLayout>
            <DefaultModalHeader>{props.checkoutName}</DefaultModalHeader>
            <DefaultModalList>

                <DefaultModalItem showOnMobile onClick={props.onSessionCreationSubmission}>
                    <CubeIcon />
                    <span>Create session</span>
                </DefaultModalItem>
                
                <DefaultModalDivider className="sm:hidden" />

                <DefaultModalItem onClick={props.onCommitMessageSelection}>
                    <AnnotationIcon />
                    <span>View commit message</span>
                </DefaultModalItem>
            </DefaultModalList>
        </DefaultModalLayout>
    </DefaultModal>;
}