import React from 'react';
import { DefaultModal } from '@/components/manager/modal/default-modal';
import { DefaultModalDivider, DefaultModalHeader, DefaultModalItem, DefaultModalLayout, DefaultModalList } from '@/components/manager/modal/default-modal-layout/default-modal-layout';
import { AnnotationIcon } from '@/components/shared/elements/icons/annotation/annotation-icon';
import { BeakerIcon } from '@/components/shared/elements/icons/beaker/beaker-icon';
import { ClipboardIcon } from '@/components/shared/elements/icons/clipboard/clipboard-icon';
import { CodeIcon } from '@/components/shared/elements/icons/code/code-icon';
import { CubeIcon } from '@/components/shared/elements/icons/cube/cube-icon';
import { LoginIcon } from '@/components/shared/elements/icons/login/login-icon';
import { TextDocumentIcon } from '@/components/shared/elements/icons/text-document/text-document-icon';
import { TrashIcon } from '@/components/shared/elements/icons/trash/trash-icon';
import { ISession } from '@/state/models';

type TProps = {
    name                   : string;
    session                : ISession;
    onCommitMessageSelect  : (session: ISession) => void;
    onEnterSessionSelect   : () => void;
    onSessionDeletionSelect: () => void;
    onCopyPermalinkSelect  : () => void;
    onShowLogsSelect       : (session: ISession) => void;
    onAPISelect            : (session: ISession) => void;
}
export const ApplicationSessionModal = (props: TProps) => {
    return <DefaultModal name={props.name}>
        <DefaultModalLayout>

            <DefaultModalHeader
                title={props.session.displayName}
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

                <DefaultModalItem onClick={() => props.onCommitMessageSelect(props.session)}>
                    <AnnotationIcon />
                    <span>View commit message</span>
                </DefaultModalItem>

                <DefaultModalItem onClick={() => props.onShowLogsSelect(props.session)}>
                    <TextDocumentIcon />
                    <span>View build logs</span>
                </DefaultModalItem>

                <DefaultModalDivider />
                
                <DefaultModalItem onClick={() => props.onAPISelect(props.session)}>
                    <CodeIcon />
                    <span>JSON API</span>
                </DefaultModalItem>

                {props.session.beingReplacedBySession && <>
                    <DefaultModalDivider />

                    <DefaultModalItem categoryHeader>
                        Replacement build
                    </DefaultModalItem>

                    <DefaultModalItem indented onClick={() => props.onCommitMessageSelect(props.session.beingReplacedBySession)}>
                        <AnnotationIcon />
                        <span>View commit message</span>
                    </DefaultModalItem>

                    <DefaultModalItem indented onClick={() => props.onShowLogsSelect(props.session.beingReplacedBySession)}>
                        <TextDocumentIcon />
                        <span>View build logs</span>
                    </DefaultModalItem>

                    <DefaultModalItem indented onClick={() => props.onAPISelect(props.session.beingReplacedBySession)}>
                        <CodeIcon />
                        <span>JSON API</span>
                    </DefaultModalItem>
                </>}

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