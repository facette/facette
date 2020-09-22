<!--
 Copyright (c) 2020, The Facette Authors

 Licensed under the terms of the BSD 3-Clause License; a copy of the license
 is available at: https://opensource.org/licenses/BSD-3-Clause
-->

<template>
    <v-modal name="chart-marker" :title="i18n.t(`labels.markers.${edit ? 'edit' : 'add'}`)" @show="onShow">
        <template v-slot="modal">
            <v-form>
                <v-label>{{ i18n.t("labels.value") }}</v-label>
                <v-input required type="number" v-autofocus v-model:value.number="marker.value"></v-input>

                <v-label>{{ i18n.t("labels.labels._", 1) }}</v-label>
                <v-input
                    :placeholder="i18n.t('labels.placeholders.default', [marker.value])"
                    v-model:value="marker.label"
                ></v-input>

                <v-label>{{ i18n.t("labels.color") }}</v-label>
                <v-color class="half" v-model:value="marker.color"></v-color>

                <v-label>{{ i18n.t("labels.charts.axes._", 1) }}</v-label>
                <v-select
                    class="half"
                    required
                    :options="axes"
                    :placeholder="i18n.t('labels.charts.axes.select')"
                    v-model:value="marker.axis"
                ></v-select>

                <template v-slot:bottom>
                    <v-button @click="modal.close(false)">
                        {{ i18n.t("labels.cancel") }}
                    </v-button>

                    <v-button :disabled="!marker.value || !marker.axis" primary @click="modal.close(marker)">
                        {{ i18n.t("labels.markers.set") }}
                    </v-button>
                </template>
            </v-form>
        </template>
    </v-modal>
</template>

<script lang="ts">
import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import {ref} from "vue";
import {useI18n} from "vue-i18n";

import {Marker} from "types/api";
import {SelectOption} from "types/ui";

export interface ModalChartMarkerParams {
    edit: boolean;
    marker: Marker;
}

const defaultMarker: Marker = {
    value: 0,
    label: "",
    color: "",
    axis: "left",
};

export default {
    setup(): Record<string, unknown> {
        const i18n = useI18n();

        const edit = ref(false);
        const marker = ref<Marker>(cloneDeep(defaultMarker));

        const axes = ref<Array<SelectOption>>([
            {label: i18n.t("labels.charts.axes.left"), value: "left"},
            {label: i18n.t("labels.charts.axes.right"), value: "right"},
        ]);

        const onShow = (params: ModalChartMarkerParams): void => {
            edit.value = params.edit;
            marker.value = merge({}, defaultMarker, params.marker);
        };

        return {
            axes,
            edit,
            i18n,
            onShow,
            marker,
        };
    },
};
</script>

<style lang="scss" scoped>
@import "../../../mixins";

.v-modal {
    ::v-deep(.v-modal-content) {
        @include content;
    }
}
</style>
