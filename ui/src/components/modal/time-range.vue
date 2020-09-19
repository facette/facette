<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="time-range" :title="i18n.t('labels.timeRange._')" @show="onShow">
        <template v-slot="modal">
            <v-form ref="form">
                <v-datetime
                    :label="i18n.t('labels.timeRange.from')"
                    v-autofocus.select
                    v-model:value="from"
                ></v-datetime>

                <v-datetime :label="i18n.t('labels.timeRange.to')" :min="from" v-model:value="to"></v-datetime>

                <template v-slot:bottom>
                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.cancel") }}
                    </v-button>

                    <v-button primary @click="modal.close(formatRange(from, to))">
                        {{ i18n.t("labels.timeRange.set") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import {ref} from "vue";
import {useI18n} from "vue-i18n";

import {TimeRange} from "types/api";

import {dateFormatDisplay} from "@/components/chart";
import {parseDate} from "@/helpers/date";

export interface ModalTimeRangeParams {
    range: TimeRange;
}

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        const from = ref("");
        const to = ref("");

        const formatRange = (from: string, to: string): TimeRange => {
            return {
                from: parseDate(from, dateFormatDisplay).toISO(),
                to: parseDate(to, dateFormatDisplay).toISO(),
            };
        };

        const onShow = (params: ModalTimeRangeParams): void => {
            from.value = from.value ? parseDate(params.range.from).toFormat(dateFormatDisplay) : "";
            to.value = to.value ? parseDate(params.range.to).toFormat(dateFormatDisplay) : "";
        };

        return {
            formatRange,
            from,
            i18n,
            onShow,
            to,
        };
    },
};
</script>

<style lang="scss" scoped>
.v-modal {
    // .v-datetime {
    //     height: 100%;
    // }

    .v-datetime + .v-datetime {
        margin-left: 2rem;
    }
}
</style>
