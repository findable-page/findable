<?php
if ( ! defined( 'WP_UNINSTALL_PLUGIN' ) ) {
	exit;
}

function findable_uninstall_site() {
	delete_post_meta_by_key( '_findable_scan' );
	delete_post_meta_by_key( '_findable_keyphrase' );
	delete_post_meta_by_key( '_findable_fixes' );
	delete_post_meta_by_key( '_findable_fixes_backup' );
}

if ( is_multisite() ) {
	foreach ( get_sites( array( 'fields' => 'ids', 'number' => 0 ) ) as $findable_site_id ) {
		switch_to_blog( $findable_site_id );
		findable_uninstall_site();
		restore_current_blog();
	}
} else {
	findable_uninstall_site();
}
