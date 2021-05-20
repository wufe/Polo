import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { DefaultModalDivider, DefaultModalHeader, DefaultModalItem, DefaultModalLayout, DefaultModalList } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { AnnotationIcon } from '@/components/shared/ui-elements/icons/annotation/annotation-icon';
import { BeakerIcon } from '@/components/shared/ui-elements/icons/beaker/beaker-icon';
import { ClipboardIcon } from '@/components/shared/ui-elements/icons/clipboard/clipboard-icon';
import { CodeIcon } from '@/components/shared/ui-elements/icons/code/code-icon';
import { CubeIcon } from '@/components/shared/ui-elements/icons/cube/cube-icon';
import { LoginIcon } from '@/components/shared/ui-elements/icons/login/login-icon';
import { TextDocumentIcon } from '@/components/shared/ui-elements/icons/text-document/text-document-icon';
import { TrashIcon } from '@/components/shared/ui-elements/icons/trash/trash-icon';
import { ISession } from '@/state/models';

type TProps = {
    name                   : string;
    session                : ISession;
    onCommitMessageSelect  : () => void;
    onEnterSessionSelect   : () => void;
    onSessionDeletionSelect: () => void;
    onCopyPermalinkSelect  : () => void;
    onShowLogsSelect       : () => void;
}
export const ApplicationSessionModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <DefaultModalLayout>

            <DefaultModalHeader
                title={props.session.checkout}
                subtitle={props.session.uuid} />

            <DefaultModalList>
                <DefaultModalItem onClick={props.onEnterSessionSelect}>
                    <LoginIcon />
                    <span>Enter the session</span>
                </DefaultModalItem>
                
                <DefaultModalDivider />
                
                <DefaultModalItem onClick={props.onCopyPermalinkSelect}>
                    <ClipboardIcon />
                    <span>Copy permalink</span>
                </DefaultModalItem>

                <DefaultModalItem onClick={props.onCommitMessageSelect}>
                    <AnnotationIcon />
                    <span>View commit message</span>
                </DefaultModalItem>

                <DefaultModalItem onClick={props.onShowLogsSelect}>
                    <TextDocumentIcon />
                    <span>View build logs</span>
                </DefaultModalItem>
                
                <DefaultModalItem notImplemented>
                    <CodeIcon />
                    <span>Advanced commands</span>
                </DefaultModalItem>

                <DefaultModalItem notImplemented>
                    <BeakerIcon />
                    <span>Custom commands</span>
                </DefaultModalItem>

                <DefaultModalDivider className="hidden" />
                
                <DefaultModalItem notImplemented>
                    <CubeIcon />
                    <span>Rebuild</span>
                </DefaultModalItem>

                <DefaultModalDivider />
                
                <DefaultModalItem onClick={() => props.onSessionDeletionSelect()}>
                    <TrashIcon />
                    <span>Delete</span>
                </DefaultModalItem>
            </DefaultModalList>
        </DefaultModalLayout>
    </DefaultModal>;
}